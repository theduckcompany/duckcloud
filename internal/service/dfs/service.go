package dfs

import (
	"errors"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = inodes.ErrInvalidPath
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

// cleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func cleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
