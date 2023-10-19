package files

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context) (afero.File, uuid.UUID, error)
	Open(ctx context.Context, fileID uuid.UUID) (afero.File, error)
	Delete(ctx context.Context, fileID uuid.UUID) error
}

func Init(dirPath string, fs afero.Fs, tools tools.Tools) (Service, error) {
	err := fs.MkdirAll(dirPath, 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the files directory: %w", err)
	}

	return NewFSService(fs, dirPath, tools)
}
