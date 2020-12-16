package model

import (
	"errors"
	"fmt"
	"github.com/utmhikari/repomaster/pkg/util"
	"os"
	"path/filepath"
)

// Config is the app cfg template
type Config struct {
	Port int  `json:"port"`
	RepoRoot string `json:"repoRoot"`
}

// check validity of config instance
func (c Config) Check() error{
	// check port
	if c.Port < 3000{
		return errors.New(fmt.Sprintf("invalid port number: %d", c.Port))
	}
	// check repo root
	if !util.ExistsPath(c.RepoRoot){
		err := os.Mkdir(c.RepoRoot, os.ModePerm)
		if err != nil{
			return err
		}
	} else if !util.IsDirectory(c.RepoRoot){
		return errors.New(
			fmt.Sprintf("repo root already exists but not a directory: %s", c.RepoRoot))
	} else{
		// check if the path is set as the root of the project
		if absWd, absWdErr := filepath.Abs("."); absWdErr != nil{
			if absRepoRoot, absRepoRootErr := filepath.Abs(c.RepoRoot); absRepoRootErr != nil{
				if relRepoRoot, relRepoRootErr := filepath.Rel(absWd, absRepoRoot); relRepoRootErr != nil{
					if relRepoRoot == "." || relRepoRoot == ""{
						return errors.New("repo root is the working directory")
					}
				}
			}
		}
	}
	return nil
}



