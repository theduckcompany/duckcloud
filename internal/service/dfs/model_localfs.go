package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes  inodes.Service
	files   files.Service
	folder  *folders.Folder
	folders folders.Service
	tasks   scheduler.Service
	clock   clock.Clock
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

func (s *LocalFS) ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	return s.inodes.Readdir(ctx, &inodes.PathCmd{Root: s.folder.RootFS(), FullName: name}, cmd)
}

func (s *LocalFS) CreateDir(ctx context.Context, name string) (*inodes.INode, error) {
	name = cleanPath(name)

	inode, err := s.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to MkdirAll: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) Remove(ctx context.Context, name string) error {
	name = cleanPath(name)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
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

func (s *LocalFS) Rename(ctx context.Context, oldName, newName string) error {
	oldName = cleanPath(oldName)
	newName = cleanPath(newName)

	current, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: oldName,
	})
	if err != nil {
		return fmt.Errorf("failed to Get: %w", err)
	}

	targetDir, targetName := path.Split(newName)

	targetFolder, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: targetDir,
	})

	existingFile, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     targetFolder.ID(),
		FullName: targetName,
	})
	if err != nil && !errors.Is(err, inodes.ErrNotFound) {
		return fmt.Errorf("failed to Get inodes: %w", err)
	}

	if existingFile != nil {
		err := s.inodes.Remove(ctx, existingFile)
	}

	return ErrNotImplemented
}

func (s *LocalFS) Get(ctx context.Context, name string) (*inodes.INode, error) {
	name = cleanPath(name)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return res, nil
}

func (s *LocalFS) Download(ctx context.Context, name string) (io.ReadSeekCloser, error) {
	inode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
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

func (s *LocalFS) Upload(ctx context.Context, name string, w io.Reader) error {
	name = cleanPath(name)

	dirPath, fileName := path.Split(name)
	if dirPath == "" {
		dirPath = "/"
	}

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: dirPath,
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
	err = s.tasks.RegisterFileUploadTask(ctx, &scheduler.FileUploadArgs{
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
