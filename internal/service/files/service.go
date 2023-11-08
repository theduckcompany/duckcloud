package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spf13/afero"
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
		// XXX:MULTI-WRITE
		//
		// This function is idempotent so no worries. If it fails the server doesn't start
		// so we are sur that it will be run again until it's completely successful.
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

func (s *FSService) Upload(ctx context.Context, r io.Reader) (uuid.UUID, error) {
	fileID := s.uuid.New()

	file, err := s.open(ctx, fileID)
	if err != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return "", fmt.Errorf("failed to open the file: %w", err)
	}

	_, err = io.Copy(file, r)
	if err != nil {
		_ = file.Close()
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return "", fmt.Errorf("failed to write the file: %w", err)
	}

	err = file.Close()
	if err != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return "", fmt.Errorf("failed to close the file: %w", err)
	}

	return fileID, nil
}

func (s *FSService) Download(ctx context.Context, fileID uuid.UUID) (io.ReadSeekCloser, error) {
	file, err := s.open(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %w", err)
	}

	return file, nil
}

func (s *FSService) Delete(ctx context.Context, fileID uuid.UUID) error {
	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	err := s.fs.Remove(filePath)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *FSService) open(ctx context.Context, fileID uuid.UUID) (afero.File, error) {
	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o600)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, errs.BadRequest(ErrNotExist)
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to open %q: %w", filePath, err))
	}

	return file, nil
}
