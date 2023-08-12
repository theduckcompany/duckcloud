package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/storage"
)

type File struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
	cmd      *inodes.PathCmd
}

func (d *File) Close() error                                 { return nil }
func (d *File) Read(p []byte) (int, error)                   { return 0, fs.ErrInvalid }
func (d *File) Write(p []byte) (int, error)                  { return 0, fs.ErrInvalid }
func (d *File) Seek(offset int64, whence int) (int64, error) { return 0, fs.ErrInvalid }

func (d *File) Readdir(count int) ([]fs.FileInfo, error) {
	if !d.inode.Mode().IsDir() {
		return nil, fs.ErrInvalid
	}

	// TODO: Check if we should use the context from `OpenFile`
	res, err := d.inodeSvc.Readdir(context.Background(), d.cmd, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": ""},
		Limit:      count,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Readdir: %w", err)
	}

	var infos []fs.FileInfo

	for idx := range res {
		infos = append(infos, &res[idx])
	}

	return infos, nil
}

func (d *File) Stat() (os.FileInfo, error) {
	return d.inode, nil
}
