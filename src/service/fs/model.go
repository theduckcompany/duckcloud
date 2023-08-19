package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

type File struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
	fileSvc  files.Service
	cmd      *inodes.PathCmd
	file     afero.File
}

func (f *File) Close() error {
	if f.file == nil {
		return nil
	}

	return f.file.Close()
}

func (f *File) Read(p []byte) (int, error) {
	var err error

	if f.file == nil {
		f.file, err = f.fileSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the file %q: %w", f.inode.ID(), err)
		}
	}

	return f.file.Read(p)
}

func (f *File) Write(p []byte) (int, error) {
	var err error

	if f.file == nil {
		f.file, err = f.fileSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the file %q: %w", f.inode.ID(), err)
		}
	}

	return f.file.Write(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	var err error

	if f.file == nil {
		f.file, err = f.fileSvc.Open(context.Background(), f.inode.ID())
		if err != nil {
			return 0, fmt.Errorf("failed to Open the file %q: %w", f.inode.ID(), err)
		}
	}

	return f.file.Seek(offset, whence)
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
