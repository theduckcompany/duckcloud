package dav

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdfs "io/fs"
	"os"

	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

var ErrConcurentReadWrite = errors.New("concurent read and write unauthorized")

type File struct {
	inode  *inodes.INode
	fs     fs.FS
	writer *io.PipeWriter
	reader *io.PipeReader
}

func NewFile(inode *inodes.INode, fs fs.FS) *File {
	return &File{inode, fs, nil, nil}
}

func (f *File) Close() error {
	var err error
	if f.writer != nil {
		err = f.writer.Close()
		f.writer = nil
	}

	if f.reader != nil {
		err = f.reader.Close()
		f.reader = nil
	}

	return err
}

func (f *File) Read(p []byte) (int, error) {
	if f.writer != nil {
		return 0, ErrConcurentReadWrite
	}

	if f.reader == nil {
		// Initialize the read pipeline at the first Read
		var w *io.PipeWriter
		f.reader, w = io.Pipe()

		file, err := f.fs.Download(context.Background(), f.inode)
		if err != nil {
			return 0, fmt.Errorf("failed to Download: %w", err)
		}

		go func(w *io.PipeWriter, fileReader io.ReadCloser) {
			defer fileReader.Close()
			_, err := io.Copy(w, fileReader)
			w.CloseWithError(err)
		}(w, file)
	}

	return f.reader.Read(p)
}

func (f *File) Write(p []byte) (int, error) {
	if f.reader != nil {
		return 0, ErrConcurentReadWrite
	}

	if f.writer == nil {
		// Initialize the write pipeline at the first Write
		var r *io.PipeReader
		r, f.writer = io.Pipe()

		go func(r *io.PipeReader) {
			err := f.fs.Upload(context.Background(), f.inode, r)
			r.CloseWithError(err)
		}(r)
	}

	return f.writer.Write(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return 0, stdfs.ErrInvalid
}

func (f *File) ReadDir(count int) ([]stdfs.DirEntry, error) {
	return nil, stdfs.ErrInvalid
}

func (f *File) Readdir(count int) ([]stdfs.FileInfo, error) {
	return nil, stdfs.ErrInvalid
}

func (f *File) Stat() (os.FileInfo, error) {
	return f.inode, nil
}
