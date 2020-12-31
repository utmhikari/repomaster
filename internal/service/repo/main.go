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
var cache sync.Map

// CacheItem cache id and its repo info
type CacheItem struct {
	ID   uint64 `json:"id"`
	Repo Repo   `json:"repo"`
}

// GetCacheSnapshot get snapshot of the cache
func GetCacheSnapshot() []CacheItem {
	var cacheSnapshot []CacheItem
	cache.Range(func(k, v interface{}) bool {
		id, idOk := k.(uint64)
		ctx, ctxOk := v.(*context)
		if !idOk || !ctxOk {
			return true
		}
		repoCopy := ctx.v
		cacheSnapshot = append(cacheSnapshot, CacheItem{
			ID:   id,
			Repo: repoCopy,
		})
		return true
	})
	sort.Slice(cacheSnapshot, func(i, j int) bool {
		return cacheSnapshot[i].ID < cacheSnapshot[j].ID
	})
	return cacheSnapshot
}

// getContext get context by id
func getContext(id uint64) *context {
	ctxInterface, ok := cache.Load(id)
	if !ok{
		return nil
	}
	ctx, ctxOk := ctxInterface.(*context)
	if !ctxOk {
		return nil
	}
	return ctx
}

// GetRepoRoot get the local root of repo by unique id
func getRepoRoot(id uint64) string {
	return filepath.Join(cfg.Global().RepoRoot, strconv.FormatUint(id, 10))
}

// GetRepo get info of repo
func GetRepo(id uint64) *Repo {
	ctx := getContext(id)
	if ctx == nil {
		return nil
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return &ctx.v
}

// FindRepoByHash get info of repo by specific url and hash
func FindRepoByHash(t Type, url string, hash string) (uint64, *Repo) {
	var repoID uint64 = 0
	var repoInst *Repo = nil
	cache.Range(func(k, v interface{}) bool {
		id, idOk := k.(uint64)
		ctx, ctxOk := v.(context)
		if !idOk || !ctxOk {
			return true
		}
		ctx.mu.RLock()
		defer ctx.mu.RUnlock()
		repoCopy := ctx.v
		if repoCopy.Status == StatusActive &&
			repoCopy.Type == t &&
			repoCopy.URL == url &&
			repoCopy.Commit.Hash == hash {
			repoID = id
			repoInst = &repoCopy
			return false
		}
		return true
	})
	return repoID, repoInst
}

// createContext create a new context by id
func createContext(id uint64, t Type, s Status) {
	cache.Store(id, &context{
		root: getRepoRoot(id),
		mu:   sync.RWMutex{},
		v: Repo{
			Type:   t,
			Status: s,
			Commit: Commit{},
		},
	})
}

// createDefaultContext create default repo context, with everything unknown
func createDefaultContext(id uint64) {
	createContext(id, TypeUnknown, StatusUnknown)
}

// requestNewContextWithID request a new context instance, with its ID
func requestNewContextWithID(t Type, s Status) (*context, uint64) {
	var i uint64 = 1
	for ; i <= math.MaxUint64; i++ {
		_, ok := cache.Load(i)
		if !ok {
			createContext(i, t, s)
			return getContext(i), i
		}
	}
	return nil, 0
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
	cache.Delete(id)
}

// Refresh refresh the repo cache
func Refresh() {
	repoRoot := cfg.Global().RepoRoot
	log.Printf("refresh repo cache from root: %s\n", repoRoot)
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
				if _, ok := cache.Load(id); !ok {
					createDefaultContext(id)
					log.Printf("created repo context %d\n", id)
				}
				existedIDs[id] = true
			}
		}
	}
	// refresh contexts
	var idsToRefresh []uint64
	var idsToDelete []uint64
	cache.Range(func(k, v interface{}) bool {
		id, idOk := k.(uint64)
		ctx, ctxOk := v.(*context)
		if !idOk || !ctxOk {
			return true
		}
		if _, ok := existedIDs[id]; ok {
			log.Printf("context %d will be refreshed...\n", id)
			idsToRefresh = append(idsToRefresh, id)
		} else if !(ctx.v.Status == StatusUpdating) {
			log.Printf("context %d will be deleted as repo is empty...\n", id)
			idsToDelete = append(idsToDelete, id)
		}
		return true
	})
	for _, id := range idsToDelete {
		go deleteContext(id)
	}
	for _, id := range idsToRefresh {
		go refreshContextByID(id)
	}
}
