package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/storage"
)

type Directory struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
	cmd      *inodes.PathCmd
}

func (d *Directory) Close() error                                 { return nil }
func (d *Directory) Read(p []byte) (int, error)                   { return 0, fs.ErrInvalid }
func (d *Directory) Write(p []byte) (int, error)                  { return 0, fs.ErrInvalid }
func (d *Directory) Seek(offset int64, whence int) (int64, error) { return 0, fs.ErrInvalid }

func (d *Directory) Readdir(count int) ([]fs.FileInfo, error) {
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

func (d *Directory) Stat() (os.FileInfo, error) {
	return d.inode, nil
}
