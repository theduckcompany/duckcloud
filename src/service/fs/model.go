package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

type File struct {
	inode  *inodes.INode
	inodes inodes.Service
	cmd    *inodes.PathCmd
	perm   fs.FileMode
	file   afero.File
}

func (f *File) Close() error {
	err := f.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close the file: %w", err)
	}

	switch f.inode {
	case nil:
		return f.createInode()
	default:
		return f.updateInode()
	}
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

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.inode.Mode().IsDir() {
		return nil, fs.ErrInvalid
	}

	// TODO: Check if we should use the context from `OpenFile`
	res, err := f.inodes.Readdir(context.Background(), f.cmd, &storage.PaginateCmd{
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

func (f *File) createInode() error {
	ctx := context.Background()

	dir, fileName := path.Split(f.cmd.FullName)
	if dir == "" {
		dir = "/"
	}

	parent, err := f.inodes.Get(ctx, &inodes.PathCmd{
		Root:     f.cmd.Root,
		UserID:   f.cmd.UserID,
		FullName: dir,
	})
	if err != nil {
		return fmt.Errorf("failed to Get: %w", err)
	}

	f.inode, err = f.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent: parent.ID(),
		UserID: f.cmd.UserID,
		Mode:   f.perm,
		Name:   fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to CreateFile: %w", err)
	}

	return nil
}

func (f *File) updateInode() error {
	// TODO: Update the accessTime, checksum, size, etc

	return nil
}
