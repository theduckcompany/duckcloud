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

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, meta *FileMeta) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileMeta, error)
	Delete(ctx context.Context, fileID uuid.UUID) error
	GetByChecksum(ctx context.Context, checksum string) (*FileMeta, error)
}

type FileService struct {
	storage Storage
	fs      afero.Fs
	uuid    uuid.Service
}

func NewFileService(storage Storage, rootFS afero.Fs, tools tools.Tools) *FileService {
	return &FileService{storage, rootFS, tools.UUID()}
}

func (s *FileService) Upload(ctx context.Context, r io.Reader) (uuid.UUID, error) {
	fileID := s.uuid.New()

	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", errs.Internal(fmt.Errorf("failed to create the file: %w", err))
	}

	_, err = io.Copy(file, r)
	if err != nil {
		_ = file.Close()
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return "", errs.Internal(fmt.Errorf("failed to write the file: %w", err))
	}

	err = file.Close()
	if err != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return "", errs.Internal(fmt.Errorf("failed to close the file: %w", err))
	}

	return fileID, nil
}

func (s *FileService) GetMetadataByChecksum(ctx context.Context, checksum string) (*FileMeta, error) {
	res, err := s.storage.GetByChecksum(ctx, checksum)
	if errors.Is(err, errNotFound) {
		return nil, ErrNotExist
	}

	return res, err
}

func (s *FileService) GetMetadata(ctx context.Context, fileID uuid.UUID) (*FileMeta, error) {
	res, err := s.storage.GetByID(ctx, fileID)
	if errors.Is(err, errNotFound) {
		return nil, ErrNotExist
	}

	return res, err
}

func (s *FileService) Download(ctx context.Context, fileID uuid.UUID) (io.ReadSeekCloser, error) {
	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_RDONLY, 0o600)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errs.BadRequest(ErrNotExist)
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to open the file: %w", err))
	}

	return file, nil
}

func (s *FileService) Delete(ctx context.Context, fileID uuid.UUID) error {
	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	err := s.storage.Delete(ctx, fileID)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to delete the file metadatas: %w", err))
	}

	err = s.fs.Remove(filePath)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}
