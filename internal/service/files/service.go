package files

import (
	"context"
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
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
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

type service struct {
	masterkey masterkey.Service
	storage   Storage
	fs        afero.Fs
	uuid      uuid.Service
	clock     clock.Clock
}

func newService(storage Storage, rootFS afero.Fs, tools tools.Tools, masterkey masterkey.Service) *service {
	return &service{masterkey, storage, rootFS, tools.UUID(), tools.Clock()}
}

func (s *service) Upload(ctx context.Context, r io.Reader) (*FileMeta, error) {
	fileID := s.uuid.New()

	idStr := string(fileID)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to create the file: %w", err))
	}

	key, err := secret.NewKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new key: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// Start the file encryption
	encryptWriter, err := sio.EncryptWriter(file, sio.Config{Key: key.Raw()})
	if err != nil {
		return nil, fmt.Errorf("failed to create the file encryption: %w", err)
	}

	// Start the hasher job
	hashReader, hashWriter := io.Pipe()
	hasher := sha256.New()
	g.Go(func() error {
		_, err := io.Copy(hasher, hashReader)
		if err != nil {
			hashReader.CloseWithError(fmt.Errorf("failed to calculate the file hash: %w", err))
		}

		return nil
	})

	// Start the mime type detection
	var mimeStr string
	mimeReader, mimeWriter := io.Pipe()
	g.Go(func() error {
		mime, err := mimetype.DetectReader(mimeReader)
		if err != nil {
			mimeReader.CloseWithError(fmt.Errorf("failed to detect the mime type: %w", err))
			return nil
		}

		mimeStr = mime.String()

		io.Copy(io.Discard, mimeReader)

		return nil
	})

	multiWrite := io.MultiWriter(mimeWriter, hashWriter, encryptWriter)

	written, err := io.Copy(multiWrite, r)
	if err != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return nil, errs.Internal(fmt.Errorf("upload error: %w", err))
	}

	_ = mimeWriter.Close()
	_ = hashWriter.Close()

	err = encryptWriter.Close()
	if err != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return nil, errs.Internal(fmt.Errorf("failed to end the file encryption: %w", err))
	}

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	ctx = context.WithoutCancel(ctx)

	checksum := base64.RawStdEncoding.Strict().EncodeToString(hasher.Sum(nil))

	existingFile, err := s.storage.GetByChecksum(ctx, checksum)
	if err != nil && !errors.Is(err, errNotFound) {
		return nil, errs.Internal(fmt.Errorf("failed to GetByChecksum: %w", err))
	}

	if existingFile != nil {
		_ = s.Delete(context.WithoutCancel(ctx), fileID)
		return existingFile, nil
	}

	// Start the key sealing
	var sealedKey *secret.SealedKey
	sealedKey, err = s.masterkey.SealKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sealed the key: %w", err)
	}

	fileMeta := FileMeta{
		id:         fileID,
		size:       uint64(written),
		mimetype:   mimeStr,
		checksum:   checksum,
		key:        sealedKey,
		uploadedAt: s.clock.Now(),
	}

	// XXX:MULTI-WRITE
	err = s.storage.Save(ctx, &fileMeta)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to save the file meta: %w", err))
	}

	return &fileMeta, nil
}

func (s *service) GetMetadataByChecksum(ctx context.Context, checksum string) (*FileMeta, error) {
	res, err := s.storage.GetByChecksum(ctx, checksum)
	if errors.Is(err, errNotFound) {
		return nil, ErrNotExist
	}

	return res, err
}

func (s *service) GetMetadata(ctx context.Context, fileID uuid.UUID) (*FileMeta, error) {
	res, err := s.storage.GetByID(ctx, fileID)
	if errors.Is(err, errNotFound) {
		return nil, ErrNotExist
	}

	return res, err
}

func (s *service) Download(ctx context.Context, fileMeta *FileMeta) (io.ReadSeekCloser, error) {
	idStr := string(fileMeta.id)
	filePath := path.Join(idStr[:2], idStr)

	file, err := s.fs.OpenFile(filePath, os.O_RDONLY, 0o600)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errs.BadRequest(fmt.Errorf("%s: %w", filePath, ErrNotExist))
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to open the file: %w", err))
	}

	rawKey, err := s.masterkey.Open(fileMeta.key)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file key: %w", err)
	}
	// rawKey, er
	reader, err := sio.DecryptReaderAt(file, sio.Config{Key: rawKey.Raw()})
	if err != nil {
		sioErr := sio.Error{}
		if errors.As(err, &sioErr) {
			return nil, fmt.Errorf("malformed encrypted data: %w", sioErr)
		}

		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return newDecReadSeeker(reader, int64(fileMeta.size), file), nil
}

func (s *service) Delete(ctx context.Context, fileID uuid.UUID) error {
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

type decReadSeeker struct {
	r      io.ReaderAt
	closer io.Closer
	off    int64
	size   int64
}

func newDecReadSeeker(r io.ReaderAt, size int64, closer io.Closer) *decReadSeeker {
	return &decReadSeeker{
		r:      r,
		off:    0,
		size:   size,
		closer: closer,
	}
}

func (s *decReadSeeker) Read(p []byte) (int, error) {
	if s.off >= s.size {
		return 0, io.EOF
	}

	n, err := s.r.ReadAt(p, s.off)

	s.off += int64(n)

	return n, err
}

func (s *decReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = s.off + offset
	case io.SeekEnd:
		abs = s.size + offset
	default:
		return 0, errors.New("bytes.Reader.Seek: invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("bytes.Reader.Seek: negative position")
	}

	s.off = abs

	return abs, nil
}

func (s *decReadSeeker) Close() error {
	return s.closer.Close()
}
