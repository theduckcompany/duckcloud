package blocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/spf13/afero"
)

type Config struct {
	Path string `json:"path"`
}

type Service interface {
	Open(ctx context.Context, inodeID uuid.UUID) (afero.File, error)
}

func Init(cfg Config, fs afero.Fs, tools tools.Tools) (Service, error) {
	err := fs.MkdirAll(filepath.Dir(cfg.Path), 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the blocks directory: %w", err)
	}

	return NewFSService(fs, cfg.Path, tools.Logger())
}
