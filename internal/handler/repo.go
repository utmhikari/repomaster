package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/utmhikari/repomaster/internal/models"
	repoService "github.com/utmhikari/repomaster/internal/service/repo"
	"log"
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

// GetByHash get repo info
func (_ *repo) GetByHash(c *gin.Context) {
	var request models.RepoGetByHashRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	if !repoService.IsValidType(request.Type) {
		ErrorMsgResponse(c, request.Type+" is not a valid repo type")
		return
	}
	_, r := repoService.FindRepoByHash(
		repoService.Type(request.Type), request.URL, request.Hash)
	if r != nil {
		SuccessDataResponse(c, *r)
		return
	}
	if !request.CreateIfNotExist {
		ErrorMsgResponse(c, "cannot find repo")
		return
	}
	// try create repo if not exist
	// TODO: support svn
	if request.Type != string(repoService.TypeGit) {
		ErrorMsgResponse(c, "unsupported repo type to create")
		return
	}
	gitRepoCreateOptions := models.GitRepoCreateOptions{URL: request.URL, Auth: request.GitAuth}
	gitCloneOptions := gitRepoCreateOptions.ToCloneOptions()
	if gitCloneOptions == nil {
		ErrorMsgResponse(c, "failed to get git clone options")
		return
	}
	revision := models.GitRevision{Hash: request.Hash}
	repoID := repoService.CreateGitRepo(gitCloneOptions, revision, true)
	if repoID == 0 {
		ErrorMsgResponse(c, "create repo failed")
		return
	}
	newRepo := repoService.GetRepo(repoID)
	if newRepo == nil {
		ErrorMsgResponse(c, "failed to get created repo instance")
		return
	}
	//if newRepo == nil ||
	//	!newRepo.IsActive() ||
	//	string(newRepo.Type) != request.Type ||
	//	newRepo.URL != request.URL ||
	//	newRepo.Commit.Hash != request.Hash {
	//	ErrorMsgResponse(c, "request info mismatches new repo")
	//}
	SuccessDataResponse(c, newRepo)
}

// GetFileInfo get file info of specific repo
func (_ *repo) GetFileInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	var request models.RepoGetFileInfoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	stat, err := repoService.GetFileInfoOfRepo(id, request.Path)
	if err != nil {
		log.Printf(err.Error())
		ErrorMsgResponse(c, "cannot get file stat")
		return
	}
	if !stat.IsDir {
		SuccessDataResponse(c, &models.RepoGetFileInfoResponse{
			IsDir:        false,
			FileInfo:     stat,
			FileInfoList: nil,
		})
		return
	}
	fileInfoList, err := repoService.GetFileInfoListOfRepo(id, request.Path)
	if err != nil {
		log.Printf(err.Error())
		ErrorMsgResponse(c, "cannot get filelist of dir")
		return
	}
	SuccessDataResponse(c, &models.RepoGetFileInfoResponse{
		IsDir:        true,
		FileInfo:     nil,
		FileInfoList: fileInfoList,
	})
}

// GetSnapshot get snapshot of the cache
func (_ *repo) GetSnapshot(c *gin.Context) {
	snapshot := repoService.GetCacheSnapshot()
	SuccessDataResponse(c, snapshot)
	return
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
	repoID := repoService.CreateGitRepo(gitOptions, request.Revision, false)
	SuccessMsgResponse(c, fmt.Sprintf("launched git clone at repo %d", repoID))
}

// UpdateGit update an existed git repo
func (_ *repo) UpdateGit(c *gin.Context) {
	var request models.GitRepoUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}
	checkUpdateErr := repoService.UpdateGitRepo(
		request.ID, request.Revision, request.Auth.ToAuthMethod(), false)
	if checkUpdateErr != nil {
		ErrorResponse(c, checkUpdateErr)
		return
	}
	SuccessMsgResponse(c, "launched checkout")
}
