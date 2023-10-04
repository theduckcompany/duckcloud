package dav

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
)

var ErrConcurentReadWrite = errors.New("concurent read and write unauthorized")

type File struct {
	name   string
	fs     dfs.FS
	writer *io.PipeWriter
	reader io.ReadSeekCloser
}

func NewFile(path string, fs dfs.FS) *File {
	return &File{path, fs, nil, nil}
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
	var err error

	if f.writer != nil {
		return 0, ErrConcurentReadWrite
	}

	if f.reader == nil {
		f.reader, err = f.fs.Download(context.Background(), f.name)
		if err != nil {
			return 0, fmt.Errorf("failed to Download: %w", err)
		}
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
			err := f.fs.Upload(context.Background(), f.name, r)
			r.CloseWithError(err)
		}(r)
	}

	return f.writer.Write(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.reader != nil {
		return f.reader.Seek(offset, whence)
	}

	return 0, fs.ErrInvalid
}

func (f *File) ReadDir(count int) ([]fs.DirEntry, error) {
	return nil, fs.ErrInvalid
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (f *File) Stat() (os.FileInfo, error) {
	return f.fs.Get(context.Background(), f.name)
}
