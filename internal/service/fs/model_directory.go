package fs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"syscall"

	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type Directory struct {
	inode         *inodes.INode
	inodes        inodes.Service
	cmd           *inodes.PathCmd
	lastDirOffset string
}

func NewDirectory(inode *inodes.INode, inodes inodes.Service, cmd *inodes.PathCmd) *Directory {
	return &Directory{inode, inodes, cmd, ""}
}

func (d *Directory) Read(p []byte) (int, error) {
	return 0, nil
}

func (d *Directory) Write(p []byte) (int, error) {
	return 0, &fs.PathError{Op: "write", Path: d.cmd.FullName, Err: syscall.EBADF}
}

func (d *Directory) Seek(offset int64, whence int) (int64, error) {
	return 0, fs.ErrInvalid
}

func (d *Directory) Readdir(count int) ([]fs.FileInfo, error) {
	// TODO: Check if we should use the context from `OpenDirectory`
	res, err := d.inodes.Readdir(context.Background(), d.cmd, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": d.lastDirOffset},
		Limit:      count,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Readdir: %w", err)
	}

	if len(res) > 0 {
		d.lastDirOffset = res[len(res)-1].Name()
	}

	var infos []fs.FileInfo

	for idx := range res {
		infos = append(infos, &res[idx])
	}

	if len(res) < count {
		return infos, io.EOF
	}

	return infos, nil
}

func (d *Directory) Stat() (fs.FileInfo, error) {
	return d.inode, nil
}

func (d *Directory) Close() error {
	return nil
}
