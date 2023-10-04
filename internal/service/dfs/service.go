package dfs

import (
	"errors"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = errors.New("invalid path")
)

type FSService struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
	uploads uploads.Service
}

func NewFSService(inodes inodes.Service, files files.Service, folders folders.Service, uploads uploads.Service) *FSService {
	return &FSService{inodes, files, folders, uploads}
}

func (s *FSService) GetFolderFS(folder *folders.Folder) FS {
	return newLocalFS(s.inodes, s.files, folder, s.folders, s.uploads)
}
