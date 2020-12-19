package models

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
)

// RepoCreateOptions options for creating a repo
type RepoCreateOptions struct {
	URL      string `json:"url" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`

}

// ToGitCloneOptions convert RepoCreateOptions to a git clone option
func (o *RepoCreateOptions) ToGitCloneOptions() *git.CloneOptions {
	gitCloneOptions := git.CloneOptions{
		URL:      o.URL,
		Progress: os.Stdout,
	}
	if len(o.Username) == 0{
		o.Username = "git"
	}
	if len(o.Key) > 0 {
		// use ssh key
		publicKey, keyErr := ssh.NewPublicKeys(o.Username, []byte(o.Key), o.Password)
		if keyErr != nil{
			log.Printf("failed to create ssh auth with key %s! %s\n",
				o.Key, keyErr.Error())
			return nil
		}
		gitCloneOptions.Auth = publicKey
	} else {
		// use username and password
		gitCloneOptions.Auth = &http.BasicAuth{
			Username: o.Username,
			Password: o.Password,
		}
	}
	return &gitCloneOptions
}

// RepoCreateRequest request for create a new repo
type RepoCreateRequest struct {
	Type    string            `json:"type" binding:"required"`
	Options RepoCreateOptions `json:"options" binding:"required"`
}
