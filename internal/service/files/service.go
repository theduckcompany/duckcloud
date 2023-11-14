package files

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sio"
	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"golang.org/x/sync/errgroup"
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
	clock   clock.Clock
}

func NewFileService(storage Storage, rootFS afero.Fs, tools tools.Tools) *FileService {
	return &FileService{storage, rootFS, tools.UUID(), tools.Clock()}
}

func (s *FileService) Upload(ctx context.Context, r io.Reader) (uuid.UUID, error) {
	fileID := s.uuid.New()

	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", errs.Internal(fmt.Errorf("failed to create the file: %w", err))
	}

	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return "", fmt.Errorf("Failed to read random data: %v", err) // add error handling
	}

	g, ctx := errgroup.WithContext(ctx)

	encryptReader, encryptWriter := io.Pipe()
	g.Go(func() error {
		if _, err = sio.Encrypt(file, encryptReader, sio.Config{Key: key[:]}); err != nil {
			return fmt.Errorf("failed to encrypt data: %v", err)
		}

		return nil
	})

	// Start the hasher job
	hashReader, hashWriter := io.Pipe()
	hasher := sha256.New()
	g.Go(func() error {
		_, err := io.Copy(hasher, hashReader)
		return err
	})

	var mimeStr string
	mimeReader, mimeWriter := io.Pipe()
	g.Go(func() error {
		mime, err := mimetype.DetectReader(mimeReader)

		mimeStr = mime.String()

		return err
	})

	multiWrite := io.MultiWriter(mimeWriter, hashWriter, encryptWriter)

	written, err := io.Copy(multiWrite, r)
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

	_ = mimeWriter.Close()
	_ = hashWriter.Close()
	_ = encryptWriter.Close()

	err = g.Wait()
	if err != nil {
		return "", err
	}

	ctx = context.WithoutCancel(ctx)

	checksum := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	existingFile, err := s.storage.GetByChecksum(ctx, checksum)
	if err != nil && !errors.Is(err, errNotFound) {
		return "", fmt.Errorf("failed to GetByChecksum: %w", err)
	}

	if existingFile != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return existingFile.ID(), nil
	}

	// XXX:MULTI-WRITE
	err = s.storage.Save(ctx, &FileMeta{
		id:         fileID,
		size:       uint64(written),
		mimetype:   mimeStr,
		checksum:   checksum,
		uploadedAt: s.clock.Now(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to save the file meta: %w", err)
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
