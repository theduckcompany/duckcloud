package dav

import (
	"io/fs"
	"os"

	"github.com/Peltoche/neurone/src/service/inodes"
)

type Directory struct {
	inode    *inodes.INode
	inodeSvc inodes.Service
}

func (d *Directory) Close() error                                 { return nil }
func (d *Directory) Read(p []byte) (int, error)                   { return 0, fs.ErrInvalid }
func (d *Directory) Write(p []byte) (int, error)                  { return 0, fs.ErrInvalid }
func (d *Directory) Seek(offset int64, whence int) (int64, error) { return 0, fs.ErrInvalid }

func (d *Directory) Readdir(count int) ([]fs.FileInfo, error) {
	return []fs.FileInfo{}, nil
}

func (d *Directory) Stat() (os.FileInfo, error) {
	return d.inode, nil
}
