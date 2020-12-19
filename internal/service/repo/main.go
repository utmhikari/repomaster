package repo

import (
	"github.com/go-git/go-git/v5"
	"github.com/utmhikari/repomaster/internal/service/cfg"
	"github.com/utmhikari/repomaster/pkg/util"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"sync"
)

// Type repo type
type Type string

const (
	TypeUnknown Type = "unknown"
	TypeGit     Type = "git"
	TypeSvn     Type = "svn"
)

// IsValidType is type value valid
func IsValidType(t string) bool {
	return t == string(TypeGit) ||
		t == string(TypeSvn)
}

// Status repo status
type Status string

const (
	StatusUnknown  Status = "unknown"
	StatusError    Status = "error"
	StatusUpdating Status = "updating"
	StatusActive   Status = "active"
)

// IsValidStatus is status value valid
func IsValidStatus(s string) bool {
	return s == string(StatusActive) ||
		s == string(StatusError) ||
		s == string(StatusUpdating)
}

// Commit repo head commit info
type Commit struct {
	Hash    string
	Ref     string
	Message string
	Author  string
	Email   string
}

// Repo the info of a spefific repo
type Repo struct {
	// URL is the url of the repo of remote
	URL string

	// Type is the type of repo
	Type Type

	// Status is the current status of repo
	Status Status

	// Desc is the description of repo
	Desc string

	// Commit is the current commmit info of repo
	Commit Commit
}

// SetStatus set status of repo instance
func (r *Repo) SetStatus(status Status) {
	switch status {
	case StatusError:
		r.Status = StatusError
		break
	case StatusActive:
		r.Status = StatusActive
		break
	case StatusUpdating:
		r.Status = StatusUpdating
		break
	default:
		r.Status = StatusUnknown
		break
	}
}

// SetStatusError
func (r *Repo) SetStatusError(errMsg string) {
	r.SetStatus(StatusError)
	r.Desc = errMsg
}

// SetType set type of repo
func (r *Repo) SetType(repoType Type) {
	switch repoType {
	case TypeGit:
		r.Type = TypeGit
		break
	case TypeSvn:
		r.Type = TypeSvn
		break
	default:
		r.Type = TypeUnknown
		break
	}
}

// mu global runtime mutex
var mu sync.RWMutex

// context the repo context in repomaster runtime
type context struct {
	// root the local root of repo
	root string
	// mu mutex to protect repo instance
	mu sync.RWMutex
	// v the repo instance
	v Repo
}

// cache stores the repo contexts
var cache = make(map[uint64]*context)

// GetRepo get info of repo
func GetRepo(id uint64) *Repo {
	mu.RLock()
	defer mu.RUnlock()
	ctx, ok := cache[id]
	if !ok {
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return &ctx.v
}

// GetRepoRoot get the local root of repo by unique id
func getRepoRoot(id uint64) string {
	return filepath.Join(cfg.Global().RepoRoot, strconv.FormatUint(id, 10))
}

// createContext create a new context by id, no lock
func createContext(id uint64, t Type, s Status) {
	cache[id] = &context{
		root: getRepoRoot(id),
		mu:   sync.RWMutex{},
		v: Repo{
			Type:   t,
			Status: s,
			Commit: Commit{},
		},
	}
}

// createDefaultContext create default repo context, with everything unknown
func createDefaultContext(id uint64) {
	createContext(id, TypeUnknown, StatusUnknown)
}

// requestNewContextWithID request a new context instance, with its ID
func requestNewContextWithID(t Type, s Status) (*context, uint64) {
	mu.Lock()
	defer mu.Unlock()
	var i uint64 = 1
	for ; i <= math.MaxUint64; i++ {
		if _, ok := cache[i]; !ok {
			createContext(i, t, s)
			return cache[i], i
		}
	}
	return nil, 0
}

// getContext get context by id
func getContext(id uint64) *context {
	mu.RLock()
	defer mu.RUnlock()
	ctx, _ := cache[id]
	return ctx
}

// refreshContextByID initialize repo context by id
func refreshContextByID(id uint64) {
	ctx := getContext(id)
	if ctx == nil {
		log.Printf("cannot refresh ctx %d as context is nil!\n", id)
		return
	}
	// ignore updating contexts
	if ctx.v.Status == StatusUpdating {
		return
	}
	// try open as git repo
	r, err := git.PlainOpen(ctx.root)
	if err != nil {
		log.Printf("cannot open ctx %d as a git repo! %s\n", id, err.Error())
	} else {
		refreshContextOfGitRepo(ctx, r)
		return
	}
}

// Refresh refresh the repo cache
func Refresh() {
	repoRoot := cfg.Global().RepoRoot
	log.Printf("refresh repo cache from root: %s\n", repoRoot)
	mu.Lock()
	defer mu.Unlock()
	// list all files in repo root
	files, filesErr := ioutil.ReadDir(repoRoot)
	if filesErr != nil {
		panic(filesErr)
	}
	// initialize contexts
	existedIDs := make(map[uint64]bool)
	for _, file := range files {
		filename := file.Name()
		id, idErr := strconv.ParseUint(filename, 10, 64)
		if idErr == nil {
			repoRoot := filepath.Join(repoRoot, filename)
			if util.IsDirectory(repoRoot) {
				if _, ok := cache[id]; !ok {
					createDefaultContext(id)
					log.Printf("created repo context %d\n", id)
				}
				existedIDs[id] = true
			}
		}
	}
	// refresh contexts
	for id, ctx := range cache {
		if _, ok := existedIDs[id]; ok {
			go refreshContextByID(id)
		} else if !(ctx != nil && ctx.v.Status == StatusUpdating) {
			log.Printf("context %d will be deleted as repo is empty...\n", id)
			delete(cache, id)
		}
	}
}
