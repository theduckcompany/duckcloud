package dfs

import (
	"errors"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

const DefaultSpaceName = "My files"

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrInvalidPath     = errors.New("invalid path")
	ErrInvalidRoot     = errors.New("invalid root")
	ErrInvalidParent   = errors.New("invalid parent")
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrIsNotDir        = errors.New("not a directory")
	ErrIsADir          = errors.New("is a directory")
	ErrNotFound        = errors.New("inode not found")
	ErrAlreadyExists   = errors.New("already exists")
)

type FSService struct {
	storage   Storage
	files     files.Service
	spaces    spaces.Service
	scheduler scheduler.Service
	tools     tools.Tools
}

func NewFSService(storage Storage, files files.Service, spaces spaces.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{storage, files, spaces, tasks, tools}
}

func (s *FSService) GetSpaceFS(space *spaces.Space) FS {
	return newLocalFS(s.storage, s.files, s.spaces, s.scheduler, s.tools)
}
