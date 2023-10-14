package dfs

import (
	"errors"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = inodes.ErrInvalidPath
)

type FSService struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
	tasks   scheduler.Service
	tools   tools.Tools
}

func NewFSService(inodes inodes.Service, files files.Service, folders folders.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{inodes, files, folders, tasks, tools}
}

func (s *FSService) GetFolderFS(folder *folders.Folder) FS {
	return newLocalFS(s.inodes, s.files, folder, s.folders, s.tasks, s.tools)
}

// cleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func cleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
