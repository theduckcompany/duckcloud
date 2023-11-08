package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Upload(ctx context.Context, r io.Reader) (uuid.UUID, error)
	Download(ctx context.Context, fileID uuid.UUID) (io.ReadSeekCloser, error)
	Delete(ctx context.Context, fileID uuid.UUID) error
}

func Init(dirPath string, fs afero.Fs, tools tools.Tools) (Service, error) {
	err := fs.MkdirAll(dirPath, 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the files directory: %w", err)
	}

	return NewFSService(fs, dirPath, tools)
}
