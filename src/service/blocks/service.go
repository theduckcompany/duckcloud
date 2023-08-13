package blocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/spf13/afero"
)

var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrDirNotExists = errors.New("dir doesn't exists")
)

type FSService struct {
	fs afero.Fs
}

func NewFSService(fs afero.Fs, rootPath string) (*FSService, error) {
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
		err = rootFS.Mkdir(dir, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("failed to Mkdir %q: %w", dir, err)
		}
	}

	return &FSService{rootFS}, nil
}

func (s *FSService) Open(ctx context.Context, inodeID uuid.UUID) (afero.File, error) {
	idStr := string(inodeID)
	filePath := path.Join(idStr[:2], idStr[2:4], string(inodeID))

	file, err := s.fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to Open %q: %w", filePath, err)
	}

	return file, nil
}
