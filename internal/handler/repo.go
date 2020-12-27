package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/utmhikari/repomaster/internal/models"
	repoService "github.com/utmhikari/repomaster/internal/service/repo"
	"strconv"
)

type repo struct{}

// Repo is the repo handler instance
var Repo repo

// GetByID get repo info by ID
func (_ *repo) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	r := repoService.GetRepo(id)
	if r == nil {
		ErrorMsgResponse(c, fmt.Sprintf("cannot get repo of id %d", id))
		return
	}
	SuccessDataResponse(c, *r)
}

// CreateGit create a new git repo
func (_ *repo) CreateGit(c *gin.Context) {
	var request models.GitRepoCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	if request.Type != string(repoService.TypeGit) {
		ErrorMsgResponse(c, fmt.Sprintf("invalid repo type %s", request.Type))
		return
	}
	gitOptions := request.Options.ToCloneOptions()
	if gitOptions == nil {
		ErrorMsgResponse(c, "cannot get clone options for git repo")
		return
	}
	repoID := repoService.CreateGitRepo(gitOptions, request.Version)
	SuccessMsgResponse(c, fmt.Sprintf("launched git clone at repo %d", repoID))
}

// UpdateGit update an existed git repo
func (_ *repo) UpdateGit(c *gin.Context){
	var request models.GitRepoUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	checkUpdateErr := repoService.UpdateGitRepo(
		request.ID, request.Version, request.Auth.ToAuthMethod())
	if checkUpdateErr != nil{
		ErrorResponse(c, checkUpdateErr)
		return
	}
	SuccessMsgResponse(c, "launched checkout")
}
