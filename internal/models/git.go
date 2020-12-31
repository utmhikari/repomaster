package models

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
)

// GitRevision specification of a git version
type GitRevision struct {
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Hash   string `json:"hash"`
}

// GitAuth auth cfg of git
type GitAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

// ToAuthMethod convert GitAuth to transport.AuthMethod
func (a *GitAuth) ToAuthMethod() transport.AuthMethod {
	if a.Username == "" && a.Password == "" && a.Key == "" {
		return nil
	}
	if a.Username == "" {
		a.Username = "git"
	}
	if len(a.Key) > 0 {
		// use ssh key
		publicKey, keyErr := ssh.NewPublicKeys(a.Username, []byte(a.Key), a.Password)
		if keyErr != nil {
			log.Printf("failed to create ssh auth with key %s! %s\n",
				a.Key, keyErr.Error())
			return nil
		}
		return publicKey
	}
	// use https
	return &http.BasicAuth{
		Username: a.Username,
		Password: a.Password,
	}
}

// GitRepoCreateOptions options for creating a repo
type GitRepoCreateOptions struct {
	URL  string  `json:"url" binding:"required"`
	Auth GitAuth `json:"auth"`
}

// ToCloneOptions convert GitRepoCreateOptions to git.CloneOptions
func (o *GitRepoCreateOptions) ToCloneOptions() *git.CloneOptions {
	gitCloneOptions := git.CloneOptions{
		URL:      o.URL,
		Progress: os.Stdout,
	}
	gitCloneOptions.Auth = o.Auth.ToAuthMethod()
	return &gitCloneOptions
}

// GitRepoCreateRequest request for create a new git repo
type GitRepoCreateRequest struct {
	Type     string               `json:"type" binding:"required"`
	Options  GitRepoCreateOptions `json:"options" binding:"required"`
	Revision GitRevision          `json:"revision"`
}

// GitRepoUpdateRequest request for update version of an existed git repo
type GitRepoUpdateRequest struct {
	ID       uint64      `json:"id" binding:"required"`
	Revision GitRevision `json:"revision"`
	Auth     GitAuth     `json:"auth"`
}
