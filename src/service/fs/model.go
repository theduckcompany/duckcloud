package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/myminicloud/myminicloud/src/service/blocks"
	"github.com/myminicloud/myminicloud/src/service/inodes"
	"github.com/myminicloud/myminicloud/src/tools/storage"
	"github.com/spf13/afero"
)

type File struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
	blockSvc blocks.Service
	cmd      *inodes.PathCmd
	block    afero.File
}

func (f *File) Close() error {
	if f.block == nil {
		return nil
	}

	return f.block.Close()
}

func (f *File) Read(p []byte) (int, error) {
	var err error

	if f.block == nil {
		f.block, err = f.blockSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the block %q: %w", f.inode.ID(), err)
		}
	}

	return f.block.Read(p)
}

func (f *File) Write(p []byte) (int, error) {
	var err error

	if f.block == nil {
		f.block, err = f.blockSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the block %q: %w", f.inode.ID(), err)
		}
	}

	return f.block.Write(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	var err error

	if f.block == nil {
		f.block, err = f.blockSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the block %q: %w", f.inode.ID(), err)
		}
	}

	return f.block.Seek(offset, whence)
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.inode.Mode().IsDir() {
		return nil, fs.ErrInvalid
	}

	// TODO: Check if we should use the context from `OpenFile`
	res, err := f.inodeSvc.Readdir(context.Background(), f.cmd, &storage.PaginateCmd{
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

func (f *File) Stat() (os.FileInfo, error) {
	return f.inode, nil
}
