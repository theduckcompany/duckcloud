package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestFileService(t *testing.T) {
	ctx := context.Background()

	t.Run("Upload and Download success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		svc := NewFileService(storage, fs, tools)

		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		assert.NoError(t, err)
		assert.NotEmpty(t, fileID)

		reader, err := svc.Download(ctx, fileID)
		assert.NoError(t, err)
		res, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)
	})

	t.Run("Upload with a fs error", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		svc := NewFileService(storage, fs, tools)

		// Create an fs error by removing the write permission
		svc.fs = afero.NewReadOnlyFs(fs)

		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		assert.Empty(t, fileID)
		assert.ErrorContains(t, err, "operation not permitted")
		assert.ErrorContains(t, err, "internal: failed to create")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		svc := NewFileService(storage, fs, tools)

		// Create a file
		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		require.NoError(t, err)
		assert.NotEmpty(t, fileID)

		// Delete it
		err = svc.Delete(ctx, fileID)
		assert.NoError(t, err)

		// Check it doesn't exists
		res, err := svc.Download(ctx, fileID)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNotExist)
	})

	t.Run("Upload with a copy error", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		svc := NewFileService(storage, fs, tools)

		// Create a file
		fileID, err := svc.Upload(ctx, iotest.ErrReader(fmt.Errorf("some-error")))
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "failed to write the file")
		assert.ErrorContains(t, err, "some-error")
		assert.Empty(t, fileID)
	})

	t.Run("GetMetadata success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		storageMock := NewMockStorage(t)
		svc := NewFileService(storageMock, fs, tools)

		storageMock.On("GetByID", mock.Anything, ExampleFile1.ID()).Return(&ExampleFile1, nil).Once()

		res, err := svc.GetMetadata(ctx, ExampleFile1.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFile1, res)
	})

	t.Run("GetMetadataByChecksum success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		storageMock := NewMockStorage(t)
		svc := NewFileService(storageMock, fs, tools)

		storageMock.On("GetByChecksum", mock.Anything, "some-checksum").Return(&ExampleFile1, nil).Once()

		res, err := svc.GetMetadataByChecksum(ctx, "some-checksum")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFile1, res)
	})
}
