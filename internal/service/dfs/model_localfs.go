package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/files"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
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

func (s *LocalFS) ListDir(ctx context.Context, dirPath string, cmd *storage.PaginateCmd) ([]INode, error) {
	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   dirPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", dirPath, err)
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
		return fmt.Errorf("failed to Get inodes: %w", err)
	}

	err = s.inodes.Remove(ctx, res)
	if err != nil {
		return fmt.Errorf("failed to Remove: %w", err)
	}

	return nil
}

func (s *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	sourceINode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   cleanPath(oldPath),
	})
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	err = s.scheduler.RegisterFSMove(ctx, &scheduler.FSMoveArgs{
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

	file, err := s.files.Open(ctx, *fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to Open file %q: %w", inode.ID(), err)
	}

	return file, nil
}

func (s *LocalFS) Upload(ctx context.Context, filePath string, w io.Reader) error {
	filePath = cleanPath(filePath)

	dirPath, fileName := path.Split(filePath)
	if dirPath == "" {
		dirPath = "/"
	}

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Folder: s.folder,
		Path:   dirPath,
	})
	if err != nil {
		return fmt.Errorf("failed to Get the dir: %w", err)
	}

	file, fileID, err := s.files.Create(ctx)
	if err != nil {
		return fmt.Errorf("failed to Create file: %w", err)
	}

	success := false
	defer func() {
		// In case of error rollback the file creation. Use a panic to
		// ensure it is executed, even in case of a panic.
		if !success {
			_ = s.files.Delete(ctx, fileID)
		}
	}()

	_, err = io.Copy(file, w)
	if err != nil {
		// In case of error rollback the file creation.
		_ = s.files.Delete(ctx, fileID)
		return errs.Internal(fmt.Errorf("failed to copy the file content: %w", err))
	}

	err = file.Close()
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to close the file: %w", err))
	}

	ctx = context.WithoutCancel(ctx)
	// XXX:MULTI-WRITE
	//
	// Once a file is uploaded and closed we need to process it in order to make
	// it available.
	err = s.scheduler.RegisterFileUploadTask(ctx, &scheduler.FileUploadArgs{
		FolderID:   s.folder.ID(),
		Directory:  dir.ID(),
		FileName:   fileName,
		FileID:     fileID,
		UploadedAt: s.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to Register the upload: %w", err)
	}

	success = true

	return nil
}
