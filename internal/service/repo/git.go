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

// refreshGitRepo refresh context
func (c *context) refreshGitRepo() {
	refreshErrPrefix := fmt.Sprintf("cannot refresh %s as git repo! ", c.root)
	// open repo
	r, err := git.PlainOpen(c.root)
	if err != nil{
		log.Printf("%s%s\n", refreshErrPrefix, err.Error())
		c.SetRepoStatusError(err.Error())
		return
	}
	// refresh data
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.v.URL == "" {
		remote, remoteErr := r.Remote("origin")
		if remoteErr != nil{
			log.Printf("failed to get remote origin! %s", remoteErr.Error())
			c.v.SetStatusError(remoteErr.Error())
			return
		}
		remoteCfg := remote.Config()
		if remoteCfg == nil{
			log.Printf("failed to get config of remote origin!")
			c.v.SetStatusError("cannot get remote origin cfg")
			return
		}
		for _, remoteURL := range remoteCfg.URLs{
			if remoteURL != ""{
				c.v.URL = remoteURL
				break
			}
		}
		if c.v.URL == ""{
			log.Printf("failed to get url from remote origin!")
			c.v.SetStatusError("cannot get remote origin url")
			return
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
		return
	}
	c.v.Commit.Ref = head.Name().String()
	headCommit, err := r.CommitObject(head.Hash())
	if err != nil {
		log.Printf("%sfailed to get head commit, %s", refreshErrPrefix, err.Error())
		c.v.SetStatusError(err.Error())
		return
	}
	c.v.Commit.Hash = headCommit.Hash.String()
	c.v.Commit.Message = headCommit.Message
	c.v.Commit.Author = headCommit.Author.Name
	c.v.Commit.Email = headCommit.Author.Email
	c.v.Type = TypeGit
	c.v.Status = StatusActive
	log.Printf("refreshed %s as git repo: %+v\n", c.root, c.v)
}

// checkoutGitRepo checkout git repo to specific version
func (c *context) checkoutGitRepo(version models.GitVersion, auth transport.AuthMethod, isNeededCleanUp bool) {
	// check current status
	if !c.IsRepoStatusNormal(){
		curStatus := c.v.Status
		log.Printf("failed to checkout repo at %s! current status is %s\n",
			c.root, string(curStatus))
		return
	}
	c.SetRepoStatus(StatusUpdating)
	log.Printf("checkout repo at %s to version %+v...\n", c.root, version)
	// init git repo worktree instance
	r, err := git.PlainOpen(c.root)
	if err != nil{
		log.Printf("failed to checkout repo at %s! cannot open repo! %s\n",
			c.root, err.Error())
		c.SetRepoStatusError(err.Error())
		return
	}
	w, err := r.Worktree()
	if err != nil{
		log.Printf("failed to get worktree of repo %s! %s\n", c.root, err.Error())
		c.SetRepoStatusError(err.Error())
		return
	}
	// check if cleanup is needed
	if isNeededCleanUp {
		log.Printf("cleaning up repo at %s...\n", c.root)
		// reset remote origin as original URL
		_ = r.DeleteRemote("origin")
		_, remoteErr := r.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{c.v.URL},
		})
		if remoteErr != nil{
			log.Printf("failed to reset remote origin of repo %s! %s\n",
				c.root, remoteErr.Error())
			c.SetRepoStatusError(remoteErr.Error())
			return
		}
		// reset --hard
		resetErr := w.Reset(&git.ResetOptions{
			Mode:   git.HardReset,
		})
		if resetErr != nil{
			log.Printf("failed to reset hard at repo %s! %s\n", c.root, resetErr.Error())
			c.SetRepoStatusError(resetErr.Error())
			return
		}
		// clean -df
		cleanErr := w.Clean(&git.CleanOptions{
			Dir: true,
		})
		if cleanErr != nil{
			log.Printf("failed to clean repo at %s! %s\n", c.root, cleanErr.Error())
			c.SetRepoStatusError(cleanErr.Error())
			return
		}
	}
	// pull newest
	log.Printf("pull repo %s from URL %s...\n", c.root, c.v.URL)
	var pullErr error = nil
	if auth == nil{
		log.Printf("warning! pulling repo %s at %s with no authentication!\n", c.root, c.v.URL)
		pullErr = w.Pull(&git.PullOptions{})
	} else{
		pullErr = w.Pull(&git.PullOptions{Auth: auth})
	}
	if pullErr != nil && pullErr != git.NoErrAlreadyUpToDate{
		log.Printf("failed to pull git repo %s --- %s\n", c.root, pullErr.Error())
		c.SetRepoStatusError(pullErr.Error())
		return
	}
	// checkout priority: commit hash > tag > branch
	// no need to set master as default branch
	var checkoutErr error = nil
	if version.Hash != ""{
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(version.Hash),
			Force: true,
		})
	} else if version.Tag != ""{
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(version.Branch),
			Force: true,
		})
	} else if version.Branch != ""{
		checkoutErr = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(version.Branch),
			Force: true,
		})
	}
	if checkoutErr != nil{
		log.Printf("failed to checkout repo at %s to version %+v! %s\n",
			c.root, version, checkoutErr.Error())
		c.SetRepoStatusError(checkoutErr.Error())
		return
	}
	log.Printf("successfully checkout repo at %s to version %+v...\n",
		c.root, version)
	// refresh info
	c.refreshGitRepo()
}

// CreateGitRepo create a new git repo, returns the context id
func CreateGitRepo(options *git.CloneOptions, version models.GitVersion) uint64 {
	if options == nil{
		return 0
	}
	log.Printf("clone git repo with params %+v...\n", options)
	// request new context with updating status, so that the context wouldn't be gced
	ctx, id := requestNewContextWithID(TypeGit, StatusUpdating)
	go func() {
		// before clone
		ctx.mu.Lock()
		ctx.v.URL = options.URL
		ctx.mu.Unlock()
		// clone
		_, err := git.PlainClone(ctx.root, false, options)
		if err != nil {
			log.Printf("failed to clone git repo %d! %s\n", id, err.Error())
			ctx.SetRepoStatusError(err.Error())
			return
		}
		log.Printf("successfully cloned git repo %d!", id)
		// checkout
		ctx.checkoutGitRepo(version, options.Auth, false)
	}()
	return id
}

// UpdateGitRepo update an existed git repo
func UpdateGitRepo(id uint64, version models.GitVersion, auth transport.AuthMethod) error {
	ctx := getContext(id)
	if ctx == nil{
		return errors.New(fmt.Sprintf("cannot get repo with ID %d", id))
	}
	go func(){
		ctx.checkoutGitRepo(version, auth, true)
	}()
	return nil
}
