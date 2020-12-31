package models

// RepoGetByHashRequest request for get repo instance by hash
type RepoGetByHashRequest struct {
	Type             string  `json:"type" binding:"required"`
	URL              string  `json:"url" binding:"required"`
	Hash             string  `json:"hash" binding:"required"`
	CreateIfNotExist bool    `json:"createIfNotExist"`
	GitAuth          GitAuth `json:"gitAuth"`
	// TODO: support svn
}

// RepoGetFileInfoRequest request for get file info from repo
type RepoGetFileInfoRequest struct {
	Path string `json:"path" binding:"required"`
}

// RepoGetFileInfoResponse response for get gile info from repo
type RepoGetFileInfoResponse struct {
	IsDir        bool        `json:"isDir"`
	FileInfo     *FileInfo   `json:"fileInfo"`
	FileInfoList *[]FileInfo `json:"fileInfoList"`
}
