package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes    inodes.Service
	files     files.Service
	folder    *folders.Folder
	folders   folders.Service
	scheduler scheduler.Service
	clock     clock.Clock
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	folder *folders.Folder,
	folders folders.Service,
	tasks scheduler.Service,
	tools tools.Tools,
) *LocalFS {
	return &LocalFS{inodes, files, folder, folders, tasks, tools.Clock()}
}

func (s *LocalFS) Folder() *folders.Folder {
	return s.folder
}

func (s *LocalFS) ListDir(ctx context.Context, path string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	path = cleanPath(path)

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   path,
	})
	if errors.Is(err, errs.ErrNotFound) {
		return nil, errs.NotFound(err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to Get inode: %w", err)
	}

	return s.inodes.Readdir(ctx, dir, cmd)
}

func (s *LocalFS) CreateDir(ctx context.Context, dirPath string) (*INode, error) {
	dirPath = cleanPath(dirPath)

	inode, err := s.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   dirPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to MkdirAll: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) Remove(ctx context.Context, path string) error {
	path = cleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   path,
	})
	if errors.Is(err, errs.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to Get inode: %w", err)
	}

	err = s.inodes.Remove(ctx, res)
	if err != nil {
		return fmt.Errorf("failed to Remove: %w", err)
	}

	return nil
}

func (s *LocalFS) Move(ctx context.Context, oldPath, newPath string) error {
	sourceINode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   cleanPath(oldPath),
	})
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	err = s.scheduler.RegisterFSMoveTask(ctx, &scheduler.FSMoveArgs{
		FolderID:    s.folder.ID(),
		SourceInode: sourceINode.ID(),
		TargetPath:  newPath,
		MovedAt:     s.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to save the task: %w", err)
	}

	return nil
}

func (s *LocalFS) Get(ctx context.Context, path string) (*INode, error) {
	path = cleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return res, nil
}

func (s *LocalFS) Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error) {
	inode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	fileID := inode.FileID()
	if fileID == nil {
		return nil, files.ErrInodeNotAFile
	}

	fileReader, err := s.files.Download(ctx, *fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to Open file %q: %w", inode.ID(), err)
	}

	return fileReader, nil
}

func (s *LocalFS) Upload(ctx context.Context, filePath string, w io.Reader) error {
	filePath = cleanPath(filePath)

	fileID, err := s.files.Upload(ctx, w)
	if err != nil {
		return fmt.Errorf("failed to Create file: %w", err)
	}

	ctx = context.WithoutCancel(ctx)
	// XXX:MULTI-WRITE
	//
	// Once a file is uploaded and closed we need to process it in order to make
	// it available.
	err = s.scheduler.RegisterFileUploadTask(ctx, &scheduler.FileUploadArgs{
		FolderID:   s.folder.ID(),
		Path:       filePath,
		FileID:     fileID,
		UploadedAt: s.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to Register the upload: %w", err)
	}

	return nil
}
