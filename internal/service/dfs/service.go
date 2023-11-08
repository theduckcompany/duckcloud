package dfs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const DefaultFolderName = "My files"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = inodes.ErrInvalidPath
)

type FSService struct {
	inodes    inodes.Service
	files     files.Service
	folders   folders.Service
	scheduler scheduler.Service
	tools     tools.Tools
}

func NewFSService(inodes inodes.Service, files files.Service, folders folders.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{inodes, files, folders, tasks, tools}
}

func (s *FSService) GetFolderFS(folder *folders.Folder) FS {
	return newLocalFS(s.inodes, s.files, folder, s.folders, s.scheduler, s.tools)
}

func (s *FSService) RemoveFS(ctx context.Context, folder *folders.Folder) error {
	rootFS, err := s.inodes.GetByID(ctx, folder.RootFS())
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("failed to Get the rootFS for %q: %w", folder.Name(), err)
	}

	// XXX:MULTI-WRITE
	//
	// TODO: Create a folderdelete task
	if rootFS != nil {
		err = s.inodes.Remove(ctx, rootFS)
		if err != nil {
			return fmt.Errorf("failed to remove the rootFS for %q: %w", folder.Name(), err)
		}
	}
	err = s.folders.Delete(ctx, folder.ID())
	if err != nil {
		return fmt.Errorf("failed to delete the folder %q: %w", folder.ID(), err)
	}

	return nil
}

func (s *FSService) CreateFS(ctx context.Context, owners []uuid.UUID) (*folders.Folder, error) {
	root, err := s.inodes.CreateRootDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to CreateRootDir: %w", err)
	}

	// XXX:MULTI-WRITE
	folder, err := s.folders.Create(ctx, &folders.CreateCmd{
		Name:   DefaultFolderName,
		Owners: owners,
		RootFS: root.ID(),
	})
	if err != nil {
		_ = s.inodes.Remove(ctx, root)

		return nil, fmt.Errorf("failed to create the folder: %w", err)
	}

	return folder, nil
}

// cleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func cleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
