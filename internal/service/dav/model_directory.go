package dav

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	ffs "github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type Directory struct {
	info          fs.FileInfo
	ffs           ffs.FS
	name          string
	lastDirOffset string
}

func NewDirectory(info fs.FileInfo, ffs ffs.FS, name string) *Directory {
	return &Directory{info, ffs, name, ""}
}

func (d *Directory) Read(p []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (d *Directory) Write(p []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (d *Directory) Seek(offset int64, whence int) (int64, error) {
	return 0, fs.ErrInvalid
}

func (d *Directory) Readdir(count int) ([]fs.FileInfo, error) {
	// TODO: Check if we should use the context from `OpenDirectory`
	res, err := d.ffs.ListDir(context.Background(), d.name, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": d.lastDirOffset},
		Limit:      count,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ListDir: %w", err)
	}

	if len(res) > 0 {
		d.lastDirOffset = res[len(res)-1].Name()
	}

	infos := []fs.FileInfo{}

	for idx := range res {
		infos = append(infos, &res[idx])
	}

	if len(res) < count {
		return infos, io.EOF
	}

	return infos, nil
}

func (d *Directory) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *Directory) Close() error {
	return nil
}
