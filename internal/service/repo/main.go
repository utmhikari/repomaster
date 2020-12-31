package repo

import (
	"github.com/utmhikari/repomaster/internal/service/cfg"
	"github.com/utmhikari/repomaster/pkg/util"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// cache stores the repo contexts
var cache = make(map[uint64]*context)

// CacheItem cache id and its repo info
type CacheItem struct {
	ID   uint64 `json:"id"`
	Repo Repo   `json:"repo"`
}

// GetCacheSnapshot get snapshot of the cache
func GetCacheSnapshot() []CacheItem {
	var cacheSnapshot []CacheItem
	for id, ctx := range cache {
		repoCopy := ctx.v
		cacheSnapshot = append(cacheSnapshot, CacheItem{
			ID:   id,
			Repo: repoCopy,
		})
	}
	sort.Slice(cacheSnapshot, func(i, j int) bool {
		return cacheSnapshot[i].ID < cacheSnapshot[j].ID
	})
	return cacheSnapshot
}

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

// FindRepoByHash get info of repo by specific url and hash
func FindRepoByHash(t Type, url string, hash string) (uint64, *Repo) {
	for id, ctx := range cache {
		// make a copy, no need to lock
		repoCopy := ctx.v
		if repoCopy.Status == StatusActive &&
			repoCopy.Type == t &&
			repoCopy.URL == url &&
			repoCopy.Commit.Hash == hash {
			return id, &repoCopy
		}
	}
	return 0, nil
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

// refreshContextByID refresh repo context by id
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
	if _, err := ctx.getGitRepo(); err == nil {
		ctx.refreshGitRepo()
	} else {
		ctx.SetRepoStatusError("unrecognized repo")
	}
}

// deleteContext delete a context
func deleteContext(id uint64) {
	mu.Lock()
	defer mu.Unlock()
	delete(cache, id)
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
			repoRootDir := filepath.Join(repoRoot, filename)
			if util.IsDirectory(repoRootDir) {
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
			go deleteContext(id)
		}
	}
}
