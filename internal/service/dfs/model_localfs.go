package dfs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes  inodes.Service
	files   files.Service
	folder  *folders.Folder
	folders folders.Service
	upload  uploads.Service
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	folder *folders.Folder,
	folders folders.Service,
	uploads uploads.Service,
) *LocalFS {
	return &LocalFS{inodes, files, folder, folders, uploads}
}

func (s *LocalFS) Folder() *folders.Folder {
	return s.folder
}

func (s *LocalFS) ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	return s.inodes.Readdir(ctx, &inodes.PathCmd{Root: s.folder.RootFS(), FullName: name}, cmd)
}

func (s *LocalFS) CreateDir(ctx context.Context, name string) (*inodes.INode, error) {
	name, err := validatePath(name)
	if err != nil {
		return nil, err
	}

	inode, err := s.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to MkdirAll: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) RemoveAll(ctx context.Context, name string) error {
	name, err := validatePath(name)
	if err != nil {
		return err
	}

	err = s.inodes.RemoveAll(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to RemoveAll: %w", err)
	}

	return nil
}

func (s *LocalFS) Rename(ctx context.Context, oldName, newName string) error {
	_, err := validatePath(oldName)
	if err != nil {
		return err
	}

	_, err = validatePath(newName)
	if err != nil {
		return err
	}

	return ErrNotImplemented
}

func (s *LocalFS) Get(ctx context.Context, name string) (*inodes.INode, error) {
	name, err := validatePath(name)
	if err != nil {
		return nil, err
	}

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
	name, err := validatePath(name)
	if err != nil {
		return err
	}

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

	defer file.Close()

	_, err = io.Copy(file, w)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to copy the file content: %w", err))
	}

	ctx = context.WithoutCancel(ctx)

	err = s.upload.Register(ctx, &uploads.RegisterUploadCmd{
		FolderID: s.folder.ID(),
		DirID:    dir.ID(),
		FileName: fileName,
		FileID:   fileID,
	})
	if err != nil {
		return fmt.Errorf("failed to Register the upload: %w", err)
	}

	return nil
}

func validatePath(name string) (string, error) {
	if name == "" {
		return ".", nil
	}

	if !fs.ValidPath(name) {
		return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	return path.Clean(name), nil
}
