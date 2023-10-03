package files

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrDirNotExists = errors.New("dir doesn't exists")
	ErrNotAFile     = errors.New("not a file")
)

type FSService struct {
	fs afero.Fs
}

func NewFSService(fs afero.Fs, rootPath string, log *slog.Logger) (*FSService, error) {
	root := path.Clean(rootPath)

	info, err := fs.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPath, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%w: open %s: it must be a directory", ErrInvalidPath, root)
	}

	log.Info(fmt.Sprintf("load files from %s", root))

	rootFS := afero.NewBasePathFs(fs, root)

	for i := 0; i < 256; i++ {
		dir := fmt.Sprintf("%02x", i)
		err = rootFS.Mkdir(dir, 0o755)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("failed to Mkdir %q: %w", dir, err)
		}
	}

	return &FSService{rootFS}, nil
}

func (s *FSService) Open(ctx context.Context, inode *inodes.INode) (afero.File, error) {
	fileID := inode.FileID()
	if fileID == nil {
		return nil, ErrNotAFile
	}

	idStr := string(*fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to Open %q: %w", filePath, err)
	}

	return file, nil
}

func (s *FSService) Delete(ctx context.Context, inode *inodes.INode) error {
	fileID := inode.FileID()
	if fileID == nil {
		return ErrNotAFile
	}

	idStr := string(*fileID)
	filePath := path.Join(idStr[:2], idStr)

	return s.fs.Remove(filePath)
}
