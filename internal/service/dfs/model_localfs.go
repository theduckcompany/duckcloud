package dfs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes  inodes.Service
	files   files.Service
	folder  *folders.Folder
	folders folders.Service
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	folder *folders.Folder,
	folders folders.Service,
) *LocalFS {
	return &LocalFS{inodes, files, folder, folders}
}

func (s *LocalFS) Folder() *folders.Folder {
	return s.folder
}

func (s *LocalFS) ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	return s.inodes.Readdir(ctx, &inodes.PathCmd{Root: s.folder.RootFS(), FullName: name}, cmd)
}

func (s *LocalFS) CreateFile(ctx context.Context, name string) (*inodes.INode, error) {
	dir, fileName := path.Split(name)
	if dir == "" {
		dir = "/"
	}

	parent, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: dir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	if parent == nil {
		return nil, &fs.PathError{Op: "createFile", Path: dir, Err: ErrInvalidPath}
	}

	file, err := s.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent: parent.ID(),
		Name:   fileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateFile: %w", err)
	}

	return file, nil
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
		return nil, fmt.Errorf("inodes mkdir error: %w", err)
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
		return nil, fmt.Errorf("failed to open inodes: %w", err)
	}

	if res == nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}

	return res, nil
}

func (s *LocalFS) Download(ctx context.Context, inode *inodes.INode) (io.ReadSeekCloser, error) {
	file, err := s.files.Open(ctx, inode)
	if err != nil {
		return nil, fmt.Errorf("failed to Open file %q: %w", inode.ID(), err)
	}

	return file, nil
}

func (s *LocalFS) Upload(ctx context.Context, inode *inodes.INode, w io.Reader) error {
	file, err := s.files.Open(ctx, inode)
	if err != nil {
		return fmt.Errorf("failed to Open file %q: %w", inode.ID(), err)
	}

	defer file.Close()

	hasher := sha256.New()

	w = io.TeeReader(w, hasher)

	sizeWrite, err := io.Copy(file, w)
	if err != nil {
		return fmt.Errorf("failed to copy the file content: %w", err)
	}

	ctx = context.WithoutCancel(ctx)

	err = s.inodes.RegisterWrite(ctx, inode, sizeWrite, hasher)
	if err != nil {
		return fmt.Errorf("failed to RegisterWrite: %w", err)
	}

	s.folder, err = s.folders.RegisterWrite(ctx, s.folder.ID(), uint64(sizeWrite))
	if err != nil {
		return fmt.Errorf("failed to RegisterWrite into folder: %w", err)
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
