package files

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

type Config struct {
	Path string `json:"path"`
}

//go:generate mockery --name Service
type Service interface {
	Open(ctx context.Context, inode *inodes.INode) (afero.File, error)
	Delete(ctx context.Context, inod *inodes.INode) error
}

func Init(cfg Config, fs afero.Fs, tools tools.Tools) (Service, error) {
	err := fs.MkdirAll(cfg.Path, 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the files directory: %w", err)
	}

	return NewFSService(fs, cfg.Path, tools.Logger())
}