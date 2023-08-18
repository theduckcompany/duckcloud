package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/myminicloud/myminicloud/src/service/inodes"
	"github.com/myminicloud/myminicloud/src/tools/storage"
)

type Dir struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
	cmd      *inodes.PathCmd
}

func (d *Dir) Close() error                                 { return nil }
func (d *Dir) Read(p []byte) (int, error)                   { return 0, fs.ErrInvalid }
func (d *Dir) Write(p []byte) (int, error)                  { return 0, fs.ErrInvalid }
func (d *Dir) Seek(offset int64, whence int) (int64, error) { return 0, fs.ErrInvalid }

func (d *Dir) Stat() (os.FileInfo, error) {
	return d.inode, nil
}

func (d *Dir) Readdir(count int) ([]fs.FileInfo, error) {
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
