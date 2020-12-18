package repo

import (
	"github.com/go-git/go-git/v5"
	"github.com/utmhikari/repomaster/internal/model"
	"github.com/utmhikari/repomaster/internal/service/cfg"
	"github.com/utmhikari/repomaster/pkg/util"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"sync"
)

// TODO: gather all into a state instance?
// State the runtime state of repo service
type State struct{
	mu sync.RWMutex
}

// mu global mutex
var mu sync.RWMutex


// context stores the repo instance
type context struct{
	mu sync.RWMutex
	v model.Repo
}

// cache stores the repo contexts
var cache = make(map[uint64]*context)


// GetRepo get info of repo
func GetRepo(id uint64) *model.Repo{
	mu.RLock()
	defer mu.RUnlock()
	ctx, ok := cache[id]
	if !ok{
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return &ctx.v
}

// GetRepoRoot get the local root of repo by unique id
func getRepoRoot(id uint64) string{
	return filepath.Join(cfg.GlobalCfg.RepoRoot, strconv.FormatUint(id, 10))
}

// newContext create a new context by id, no lock
func newContext(id uint64) {
	cache[id] = &context{
		mu: sync.RWMutex{},
		v: model.Repo{
			ID:     id,
			Type:   model.RepoTypeUnknown,
			Status: model.RepoStatusUnknown,
			Root:   getRepoRoot(id),
			Commit: model.RepoCommit{},
		},
	}
}

// newContextID new a context id
func newContextWithID() (*context, uint64) {
	mu.Lock()
	defer mu.Unlock()
	var i uint64 = 1
	for ; i <= math.MaxUint64; i++{
		if _, ok := cache[i]; !ok{
			newContext(i)
			return cache[i], i
		}
	}
	return nil, 0
}

// getContext get context by id
func getContext(id uint64) *context{
	mu.RLock()
	defer mu.RUnlock()
	ctx, _ := cache[id]
	return ctx
}

// refreshContextOfGitRepo refresh context on git.Repository
func refreshContextOfGitRepo(ctx *context, r *git.Repository){
	if ctx == nil || r == nil{
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	head, headErr := r.Head()
	if headErr != nil || head == nil{
		var errMsg string = "empty head"
		if headErr != nil{
			errMsg = headErr.Error()
		}
		log.Printf("failed to get head! %s", errMsg)
		ctx.v.SetStatusError(errMsg)
		return
	}
	ctx.v.Commit.Ref = head.Name().String()
		headCommit, err := r.CommitObject(head.Hash())
	if err != nil{
		log.Printf("failed to get head commit! %s", err.Error())
		ctx.v.SetStatusError(err.Error())
		return
	}
	ctx.v.Commit.Hash = headCommit.Hash.String()
	ctx.v.Commit.Message = headCommit.Message
	ctx.v.Commit.Author = headCommit.Author.Name
	ctx.v.Commit.Email = headCommit.Author.Email
	ctx.v.Type = model.RepoTypeGit
	ctx.v.Status = model.RepoStatusActive
	log.Printf("Refreshed git repo %d: %+v\n", ctx.v.ID, ctx.v)
}

// NewGitRepo create a new git repo
func NewGitRepo(options git.CloneOptions) error{
	log.Printf("clone git repo with params %+v...\n", options)
	ctx, id := newContextWithID()
	go func(){
		// before clone
		ctx.mu.Lock()
		ctx.v.URL = options.URL
		ctx.v.Status = model.RepoStatusUpdating
		ctx.v.Type = model.RepoTypeGit
		ctx.mu.Unlock()
		// clone
		r, err := git.PlainClone(ctx.v.Root, false, &options)
		if err != nil{
			log.Printf("failed to clone git repo %d! %s\n", id, err.Error())
			ctx.mu.Lock()
			ctx.v.SetStatusError(err.Error())
			ctx.v.SetStatusError(err.Error())
			ctx.mu.Unlock()
			return
		}
		// after cone
		log.Printf("successfully cloned git repo %d!", id)
		refreshContextOfGitRepo(ctx, r)
	}()
	return nil
}


// refreshContextByID initialize repo context by id
func refreshContextByID(id uint64){
	ctx := getContext(id)
	if ctx == nil{
		log.Printf("cannot refresh ctx %d as context is nil!\n", id)
		return
	}
	// try open as git repo
	r, err := git.PlainOpen(ctx.v.Root)
	if err != nil{
		log.Printf("cannot open ctx %d as a git repo! %s\n", id, err.Error())
	} else{
		refreshContextOfGitRepo(ctx, r)
		return
	}
}

// Refresh refresh the repo cache
func Refresh(){
	repoRoot := cfg.GlobalCfg.RepoRoot
	log.Printf("refresh repo cache from root: %s\n", cfg.GlobalCfg.RepoRoot)
	mu.Lock()
	defer mu.Unlock()
	// list all files in repo root
	files, filesErr := ioutil.ReadDir(repoRoot)
	if filesErr != nil{
		panic(filesErr)
	}
	// initialize contexts
	for _, file := range files{
		filename := file.Name()
		id, idErr := strconv.ParseUint(filename, 10, 64)
		if idErr == nil{
			repoRoot := filepath.Join(repoRoot, filename)
			if util.IsDirectory(repoRoot){
				newContext(id)
				log.Printf("initialized repo context %d\n", id)
			}
		}
	}
	// refresh contexts
	for id, _ := range cache{
		go refreshContextByID(id)
	}
}


