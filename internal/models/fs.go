package models

import "os"

// FileInfo
type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mode  uint32 `json:"mode"`
	IsDir bool   `json:"isDir"`
}

// NewFileInfoFromStat
func NewFileInfoFromStat(fileInfo os.FileInfo) *FileInfo {
	return &FileInfo{
		Name:  fileInfo.Name(),
		Size:  fileInfo.Size(),
		Mode:  uint32(fileInfo.Mode()),
		IsDir: fileInfo.IsDir(),
	}
}
