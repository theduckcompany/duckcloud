package files

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type Config struct {
	Path string `json:"path"`
}

type Service interface {
	Open(ctx context.Context, inodeID uuid.UUID) (afero.File, error)
}

func Init(cfg Config, fs afero.Fs, tools tools.Tools) (Service, error) {
	err := fs.MkdirAll(cfg.Path, 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the files directory: %w", err)
	}

	return NewFSService(fs, cfg.Path, tools.Logger())
}
