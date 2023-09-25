package fs

import (
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

//go:generate mockery --name Service
type Service interface {
	GetFolderFS(folder *folders.Folder) FS
}

func Init(inodes inodes.Service, files files.Service, folders folders.Service) Service {
	return NewFSService(inodes, files, folders)
}
