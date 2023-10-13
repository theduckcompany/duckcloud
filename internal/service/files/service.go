package files

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrInvalidPath   = errors.New("invalid path")
	ErrInodeNotAFile = errors.New("inode doesn't point to a file")
	ErrNotExist      = errors.New("file not exists")
)

type FSService struct {
	fs   afero.Fs
	uuid uuid.Service
}

func NewFSService(fs afero.Fs, rootPath string, tools tools.Tools) (*FSService, error) {
	root := path.Clean(rootPath)

	info, err := fs.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPath, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%w: open %s: it must be a directory", ErrInvalidPath, root)
	}

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

	return &FSService{rootFS, tools.UUID()}, nil
}

func (s *FSService) Create(ctx context.Context) (afero.File, uuid.UUID, error) {
	fileID := s.uuid.New()

	file, err := s.Open(ctx, fileID)
	if err != nil {
		return nil, "", errs.Internal(fmt.Errorf("failed to open the file: %w", err))
	}

	return file, fileID, nil
}

func (s *FSService) Open(ctx context.Context, fileID uuid.UUID) (afero.File, error) {
	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o600)
	if os.IsNotExist(err) {
		return nil, errs.BadRequest(ErrNotExist)
	}
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Open %q: %w", filePath, err))
	}

	return file, nil
}

func (s *FSService) Delete(ctx context.Context, inode *inodes.INode) error {
	fileID := inode.FileID()
	if fileID == nil {
		return errs.BadRequest(ErrInodeNotAFile)
	}

	idStr := string(*fileID)
	filePath := path.Join(idStr[:2], idStr)

	err := s.fs.Remove(filePath)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}
