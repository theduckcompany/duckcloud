package dav

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/net/webdav"
)

type FSService struct {
}

func NewFSService() *FSService {
	return &FSService{}
}

func (s *FSService) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	fmt.Printf("Mkdir: %q\n\n", name)
	return webdav.ErrNotImplemented
}
func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	fmt.Printf("Open file: %q\n\n", name)
	return nil, webdav.ErrNotImplemented
}

func (s *FSService) RemoveAll(ctx context.Context, name string) error {
	fmt.Printf("Remove All: %q\n\n", name)
	return webdav.ErrNotImplemented
}

func (s *FSService) Rename(ctx context.Context, oldName, newName string) error {
	fmt.Printf("Rename %q -> %q: \n\n", oldName, newName)
	return webdav.ErrNotImplemented
}

func (s *FSService) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fmt.Printf("Stats: %q\n\n", name)
	return nil, fs.ErrNotExist
}