package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"go.uber.org/fx"
)

//go:generate mockery --name Service
type Service interface {
	Upload(ctx context.Context, r io.Reader) (uuid.UUID, error)
	Download(ctx context.Context, file *FileMeta) (io.ReadSeekCloser, error)
	Delete(ctx context.Context, fileID uuid.UUID) error
	GetMetadata(ctx context.Context, fileID uuid.UUID) (*FileMeta, error)
}

type Result struct {
	fx.Out
	Service Service
}

func Init(
	masterkey masterkey.Service,
	dirPath string,
	fs afero.Fs,
	tools tools.Tools,
	db *sql.DB,
) (Result, error) {
	storage := newSqlStorage(db)

	root := path.Clean(path.Join(dirPath, "files"))
	err := fs.MkdirAll(root, 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return Result{}, fmt.Errorf("failed to create the files directory: %w", err)
	}

	rootFS := afero.NewBasePathFs(fs, root)
	err = setupFileDirectory(rootFS)
	if err != nil {
		return Result{}, fmt.Errorf("failed to setup the file storage directory: %w", err)
	}

	service := NewFileService(storage, rootFS, tools, masterkey)

	return Result{
		Service: service,
	}, nil
}

func setupFileDirectory(rootFS afero.Fs) error {
	for i := 0; i < 256; i++ {
		dir := fmt.Sprintf("%02x", i)
		// XXX:MULTI-WRITE
		//
		// This function is idempotent so no worries. If it fails the server doesn't start
		// so we are sur that it will be run again until it's completely successful.
		err := rootFS.Mkdir(dir, 0o755)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to Mkdir %q: %w", dir, err)
		}
	}

	return nil
}
