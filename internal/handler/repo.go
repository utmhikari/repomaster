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
		Error(c, Response{
			Message: fmt.Sprintf("cannot get repo of id %d", id),
		})
		return
	}
	SuccessResponse(c, *r)
}

// Create create a new repo
func (_ *repo) Create(c *gin.Context) {
	var request models.RepoCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	if !repoService.IsValidType(request.Type) {
		Error(c, Response{
			Message: fmt.Sprintf("invalid repo type %s", request.Type),
		})
		return
	}
	if request.Type == string(repoService.TypeGit) {
		gitOptions := request.Options.ToGitCloneOptions()
		if gitOptions == nil {
			Error(c, Response{
				Message: "cannot get clone options for git repo",
			})
			return
		}
		repoID := repoService.CreateGitRepo(*gitOptions)
		Success(c, Response{
			Message: fmt.Sprintf("launched git clone at repo %d", repoID),
		})
		return
	}
	Error(c, Response{
		Message: fmt.Sprintf("unsupported repo type %s", request.Type),
	})
}
