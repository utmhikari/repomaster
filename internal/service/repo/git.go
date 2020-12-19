package repo

import (
	"github.com/go-git/go-git/v5"
	"log"
)

// refreshContextOfGitRepo refresh context on git.Repository
func refreshContextOfGitRepo(ctx *context, r *git.Repository) {
	if ctx == nil || r == nil {
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	head, headErr := r.Head()
	if headErr != nil || head == nil {
		errMsg := "head is empty"
		if headErr != nil {
			errMsg = headErr.Error()
		}
		log.Printf("failed to get head at %s! %s", ctx.root, errMsg)
		ctx.v.SetStatusError(errMsg)
		return
	}
	ctx.v.Commit.Ref = head.Name().String()
	headCommit, err := r.CommitObject(head.Hash())
	if err != nil {
		log.Printf("failed to get head commit at %s! %s", ctx.root, err.Error())
		ctx.v.SetStatusError(err.Error())
		return
	}
	ctx.v.Commit.Hash = headCommit.Hash.String()
	ctx.v.Commit.Message = headCommit.Message
	ctx.v.Commit.Author = headCommit.Author.Name
	ctx.v.Commit.Email = headCommit.Author.Email
	ctx.v.Type = TypeGit
	ctx.v.Status = StatusActive
	log.Printf("Refreshed git repo at %s: %+v\n", ctx.root, ctx.v)
}

// CreateGitRepo create a new git repo, returns the context id
func CreateGitRepo(options git.CloneOptions) uint64 {
	log.Printf("clone git repo with params %+v...\n", options)
	// request new context with updating status, so that the context wouldn't be gced
	ctx, id := requestNewContextWithID(TypeGit, StatusUpdating)
	go func() {
		// before clone
		ctx.mu.Lock()
		ctx.v.URL = options.URL
		ctx.mu.Unlock()
		// clone
		r, err := git.PlainClone(ctx.root, false, &options)
		if err != nil {
			log.Printf("failed to clone git repo %d! %s\n", id, err.Error())
			ctx.mu.Lock()
			ctx.v.SetStatusError(err.Error())
			ctx.mu.Unlock()
			return
		}
		// after cone
		log.Printf("successfully cloned git repo %d!", id)
		refreshContextOfGitRepo(ctx, r)
	}()
	return id
}
