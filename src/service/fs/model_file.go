package fs

import (
	"io/fs"
	"os"
	"syscall"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
)

type File struct {
	inode  *inodes.INode
	inodes inodes.Service
	files  files.Service
	cmd    *inodes.PathCmd
	file   afero.File
}

func (f *File) Close() error {
	if f.file == nil {
		return nil
	}

	return f.file.Close()
}

func (f *File) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *File) Write(p []byte) (int, error) {
	return f.file.Write(p)
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
	return f.inode, nil
}
