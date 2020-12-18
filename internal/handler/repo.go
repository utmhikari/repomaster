package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	repoService "github.com/utmhikari/repomaster/internal/service/repo"
	"strconv"
)

type repo struct{}

// Repo is the repo handler instance
var Repo repo

// GetRepoInfoByID get repo info by ID
func (_ *repo) GetRepoInfoByID(c *gin.Context){
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil{
		Error(c, Response{
			Message: err.Error(),
		})
		return
	}
	r := repoService.GetRepo(id)
	if r == nil{
		Error(c, Response{
			Message: fmt.Sprintf("cannot get repo of id %d", id),
		})
		return
	}
	Success(c, Response{
		Data: *r,
	})
}
