package repo

import (
	"github.com/utmhikari/repomaster/internal/models"
	"github.com/utmhikari/repomaster/pkg/util"
	"path"
)

// GetFileInfoListOfRepo list files of specific repo in specific path
func GetFileInfoListOfRepo(id uint64, dirPath string) (*[]models.FileInfo, error) {
	root := getRepoRoot(id)
	relDirPath := path.Join(root, dirPath)
	fileInfoList, err := util.ListFilesOfDirectory(relDirPath)
	if err != nil {
		return nil, err
	}
	var files []models.FileInfo
	for _, fileInfo := range *fileInfoList {
		files = append(files, *models.NewFileInfoFromStat(fileInfo))
	}
	return &files, nil
}

// GetFileInfoOfRepo get specific file stat of repo
func GetFileInfoOfRepo(id uint64, filePath string) (*models.FileInfo, error) {
	root := getRepoRoot(id)
	relFilePath := path.Join(root, filePath)
	fileInfo, err := util.GetFileStat(relFilePath)
	if err != nil {
		return nil, err
	}
	return models.NewFileInfoFromStat(fileInfo), nil
}
