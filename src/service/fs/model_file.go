package fs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io/fs"
	"os"
	"syscall"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
)

const maxBufSize = 1024 * 1024

type File struct {
	inode  *inodes.INode
	inodes inodes.Service
	files  files.Service
	cmd    *inodes.PathCmd
	file   afero.File
	buffer *bytes.Buffer
	hasher hash.Hash
}

func NewFile(inode *inodes.INode,
	inodes inodes.Service,
	files files.Service,
	cmd *inodes.PathCmd,
	file afero.File,
) *File {
	buffer := new(bytes.Buffer)
	buffer.Grow(maxBufSize)

	return &File{inode, inodes, files, cmd, file, buffer, sha256.New()}
}

func (f *File) Close() error {
	if f.file == nil {
		return nil
	}

	err := f.Sync()
	if err != nil {
		return err
	}

	return f.file.Close()
}

func (f *File) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *File) Write(p []byte) (int, error) {
	pLen, err := f.buffer.Write(p)
	if err != nil {
		return 0, err
	}

	_, err = f.hasher.Write(p)
	if err != nil {
		return 0, err
	}

	if f.buffer.Len() > maxBufSize {
		err = f.Sync()
		if err != nil {
			return 0, err
		}
	}

	return pLen, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// ReadDir function is a [fs.ReadDirFile] implementation.
func (f *File) ReadDir(count int) ([]fs.DirEntry, error) {
	return []fs.DirEntry{}, &fs.PathError{Op: "readdirent", Path: f.cmd.FullName, Err: syscall.ENOTDIR}
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	return []fs.FileInfo{}, &fs.PathError{Op: "readdirent", Path: f.cmd.FullName, Err: syscall.ENOTDIR}
}

func (f *File) Stat() (os.FileInfo, error) {
	err := f.Sync()
	if err != nil {
		return nil, err
	}

	return f.inode, nil
}

func (f *File) Sync() error {
	if f.buffer.Len() == 0 {
		return nil
	}

	sizeWrite, err := f.file.Write(f.buffer.Bytes())
	if err != nil {
		return err
	}

	err = f.inodes.RegisterWrite(context.Background(), f.inode, sizeWrite, f.hasher)
	if err != nil {
		return fmt.Errorf("failed to RegisterWrite: %w", err)
	}

	f.buffer.Reset()

	return nil
}
