package repo

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/utmhikari/repomaster/internal/models"
	"log"
)

// DefaultGitRemote origin
var DefaultGitRemote = "origin"

// getGitRepo
func (c *context) getGitRepo() (*git.Repository, error) {
	if c.root == "" {
		return nil, errors.New("cannot get git repo root")
	}
	gitRepo, err := git.PlainOpen(c.root)
	if err != nil {
		return nil, err
	}
	return gitRepo, nil
}

// refreshGitRepo refresh context
func (c *context) refreshGitRepo() bool {
	refreshErrPrefix := fmt.Sprintf("cannot refresh %s as git repo! ", c.root)
	// open repo
	r, err := c.getGitRepo()
	if err != nil {
		log.Printf("%s%s\n", refreshErrPrefix, err.Error())
		c.SetRepoStatusError(err.Error())
		return false
	}
	// refresh data
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.v.URL == "" {
		remote, remoteErr := r.Remote(DefaultGitRemote)
		if remoteErr != nil {
			log.Printf("failed to get remote origin! %s", remoteErr.Error())
			c.v.SetStatusError(remoteErr.Error())
			return false
		}
		remoteCfg := remote.Config()
		if remoteCfg == nil {
			log.Printf("failed to get config of remote!")
			c.v.SetStatusError("cannot get remote cfg")
			return false
		}
		for _, remoteURL := range remoteCfg.URLs {
			if remoteURL != "" {
				c.v.URL = remoteURL
				break
			}
		}
		if c.v.URL == "" {
			log.Printf("failed to get url from remote!")
			c.v.SetStatusError("cannot get remote url")
			return false
		}
	}
	head, headErr := r.Head()
	if headErr != nil || head == nil {
		errMsg := "head is empty"
		if headErr != nil {
			errMsg = headErr.Error()
		}
		log.Printf("%sfailed to get head, %s", refreshErrPrefix, errMsg)
		c.v.SetStatusError(errMsg)
		return false
	}
	c.v.Commit.Ref = head.Name().String()
	headCommit, err := r.CommitObject(head.Hash())
	if err != nil {
		log.Printf("%sfailed to get head commit, %s", refreshErrPrefix, err.Error())
		c.v.SetStatusError(err.Error())
		return false
	}
	c.v.Commit.Hash = headCommit.Hash.String()
	c.v.Commit.Message = headCommit.Message
	c.v.Commit.Author = headCommit.Author.Name
	c.v.Commit.Email = headCommit.Author.Email
	c.v.Type = TypeGit
	c.v.Status = StatusActive
	log.Printf("refreshed %s as git repo: %+v\n", c.root, c.v)
	return true
}

// checkoutGitRepo checkout git repo to specific revision
func (c *context) checkoutGitRepo(
	revision models.GitRevision, auth transport.AuthMethod, isNeededCleanUp bool) bool {
	// check current status
	if !c.IsRepoStatusNormal() {
		curStatus := c.v.Status
		log.Printf("failed to checkout repo at %s! current status is %s\n",
			c.root, string(curStatus))
		return false
	}
	c.SetRepoStatus(StatusUpdating)
	log.Printf("checkout repo at %s to revision %+v...\n", c.root, revision)
	// init git repo worktree instance
	r, err := git.PlainOpen(c.root)
	if err != nil {
		log.Printf("failed to checkout repo at %s! cannot open repo! %s\n",
			c.root, err.Error())
		c.SetRepoStatusError(err.Error())
		return false
	}
	w, err := r.Worktree()
	if err != nil {
		log.Printf("failed to get worktree of repo %s! %s\n", c.root, err.Error())
		c.SetRepoStatusError(err.Error())
		return false
	}
	defer c.refreshGitRepo()
	// check if cleanup is needed
	if isNeededCleanUp {
		log.Printf("cleaning up repo at %s...\n", c.root)
		// reset remote origin
		_ = r.DeleteRemote(DefaultGitRemote)
		_, remoteErr := r.CreateRemote(&config.RemoteConfig{
			Name: DefaultGitRemote,
			URLs: []string{c.v.URL},
		})
		if remoteErr != nil {
			log.Printf("failed to reset remote of repo %s! %s\n",
				c.root, remoteErr.Error())
			return false
		}
		log.Printf("successfully reset remote %s of repo %s\n", DefaultGitRemote, c.root)
		// reset --hard
		var resetErr error = nil
		head, headErr := r.Head()
		if headErr != nil {
			log.Printf("warning, cannot get head ref of repo %s\n", c.root)
			resetErr = w.Reset(&git.ResetOptions{
				Mode: git.HardReset,
			})
		} else {
			resetErr = w.Reset(&git.ResetOptions{
				Commit: head.Hash(),
				Mode: git.HardReset,
			})
		}
		if resetErr != nil {
			log.Printf("failed to reset hard at repo %s! %s\n", c.root, resetErr.Error())
			return false
		}
		log.Printf("successfully reset hard at repo %s\n", c.root)
		// clean -df
		// TODO: clean all? reset all?
		cleanErr := w.Clean(&git.CleanOptions{
			Dir: true,
		})
		if cleanErr != nil {
			log.Printf("failed to clean repo at %s! %s\n", c.root, cleanErr.Error())
			return false
		}
		log.Printf("successfully cleaned files at repo %s\n", c.root)
	}
	// pull newest
	log.Printf("pull repo %s from URL %s...\n", c.root, c.v.URL)
	var pullErr error = nil
	if auth == nil {
		log.Printf("warning! pulling repo %s from %s with no authentication!\n", c.root, c.v.URL)
		pullErr = w.Pull(&git.PullOptions{
			RemoteName: DefaultGitRemote,
		})
	} else {
		pullErr = w.Pull(&git.PullOptions{
			RemoteName: DefaultGitRemote,
			Auth: auth,
		})
	}
	if pullErr != nil && pullErr != git.NoErrAlreadyUpToDate {
		log.Printf("failed to pull git repo %s --- %s\n", c.root, pullErr.Error())
		return false
	} else {
		log.Printf("pull repo %s successfully\n", c.root)
	}
	// checkout priority: commit hash > tag > branch
	// no need to set master as default branch
	// TODO: the value of tag/branch? sliced commit hash?
	var checkoutErr error = nil
	if revision.Hash != "" {
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(revision.Hash),
			Force: true,
		})
	} else if revision.Tag != "" {
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(revision.Branch),
			Force:  true,
		})
	} else if revision.Branch != "" {
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(revision.Branch),
			Force:  true,
		})
	}
	if checkoutErr != nil {
		log.Printf("failed to checkout repo at %s to revision %+v! %s\n",
			c.root, revision, checkoutErr.Error())
		return false
	}
	log.Printf("successfully checkout repo at %s to revision %+v...\n",
		c.root, revision)
	// refresh info
	return true
}

// createGitRepo create git repo
func createGitRepo(ctx *context, options *git.CloneOptions, revision models.GitRevision) bool {
	// before clone
	ctx.mu.Lock()
	ctx.v.URL = options.URL
	ctx.mu.Unlock()
	// clone
	_, err := git.PlainClone(ctx.root, false, options)
	if err != nil {
		log.Printf("failed to clone git repo to %s --- %s", ctx.root, err.Error())
		ctx.SetRepoStatusError(err.Error())
		return false
	}
	log.Printf("successfully cloned git repo to %s", ctx.root)
	// checkout
	return ctx.checkoutGitRepo(revision, options.Auth, false)
}

// CreateGitRepo create a new git repo, returns the context id
func CreateGitRepo(options *git.CloneOptions, revision models.GitRevision, isSync bool) uint64 {
	if options == nil {
		return 0
	}
	log.Printf("clone git repo with params %+v...\n", options)
	// request new context with updating status, so that the context wouldn't be gced
	ctx, id := requestNewContextWithID(TypeGit, StatusUpdating)
	// TODO: trace clone/pull/checkout progress
	if isSync {
		if !createGitRepo(ctx, options, revision) {
			// create failed
			return 0
		}
	} else {
		go createGitRepo(ctx, options, revision)
	}
	return id
}

// UpdateGitRepo update an existed git repo
func UpdateGitRepo(id uint64, revision models.GitRevision, auth transport.AuthMethod, isSync bool) error {
	ctx := getContext(id)
	if ctx == nil {
		return errors.New(fmt.Sprintf("cannot get repo with ID %d", id))
	}
	if isSync {
		if !ctx.checkoutGitRepo(revision, auth, true) {
			return errors.New("checkout git repo failed")
		}
	} else {
		go func() {
			ctx.checkoutGitRepo(revision, auth, true)
		}()
	}
	return nil
}
