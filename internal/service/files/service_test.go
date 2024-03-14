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
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestFileService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	masterPassword := secret.NewText("1superStrongPa$$word!")

	t.Run("Upload and Download success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storage, fs, tools, masterkeySvc)

		// Run
		fileMeta, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))

		// Asserts
		require.NoError(t, err)
		assert.NotNil(t, fileMeta)

		// Run 2
		reader, err := svc.Download(ctx, fileMeta)

		// Asserts 2
		require.NoError(t, err)
		res, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)
	})

	t.Run("Upload with a fs error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storage, fs, tools, masterkeySvc)

		// Create an fs error by removing the write permission
		svc.fs = afero.NewReadOnlyFs(fs)

		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		assert.Empty(t, fileID)
		require.ErrorContains(t, err, "operation not permitted")
		require.ErrorContains(t, err, "internal: failed to create")
	})

	t.Run("Delete success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storage, fs, tools, masterkeySvc)

		// Create a file
		fileMeta, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		require.NoError(t, err)
		assert.NotNil(t, fileMeta)

		// Delete it
		err = svc.Delete(ctx, fileMeta.ID())
		require.NoError(t, err)

		// Check it doesn't exists
		res, err := svc.Download(ctx, fileMeta)
		assert.Nil(t, res)
		require.ErrorIs(t, err, ErrNotExist)
	})

	t.Run("Upload with a copy error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storage, fs, tools, masterkeySvc)

		// Create a file
		fileID, err := svc.Upload(ctx, iotest.ErrReader(fmt.Errorf("some-error")))
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "upload error")
		require.ErrorContains(t, err, "some-error")
		assert.Empty(t, fileID)
	})

	t.Run("GetMetadata success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storageMock := newMockStorage(t)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storageMock, fs, tools, masterkeySvc)

		// Data
		fileMeta := NewFakeFile(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, fileMeta.ID()).Return(fileMeta, nil).Once()

		// Run
		res, err := svc.GetMetadata(ctx, fileMeta.ID())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, fileMeta, res)
	})

	t.Run("GetMetadataByChecksum success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storageMock := newMockStorage(t)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storageMock, fs, tools, masterkeySvc)

		// Data
		fileMeta := NewFakeFile(t).Build()

		// Mocks
		storageMock.On("GetByChecksum", mock.Anything, fileMeta.checksum).Return(fileMeta, nil).Once()

		// Run
		res, err := svc.GetMetadataByChecksum(ctx, fileMeta.checksum)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, fileMeta, res)
	})

	t.Run("Download an invalid content", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := sqlstorage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, tools)
		require.NoError(t, err)
		masterkeySvc.GenerateMasterKey(ctx, &masterPassword)
		require.NoError(t, err)
		svc := newService(storage, fs, tools, masterkeySvc)

		// Data
		content := []byte("not encrypted")
		fileMeta := NewFakeFile(t).WithContent(content).Build()
		err = afero.WriteFile(fs, fmt.Sprintf("%s/%s", fileMeta.id[:2], fileMeta.id), content, 0o755)
		require.NoError(t, err)

		// Run
		reader, err := svc.Download(ctx, fileMeta)

		// Asserts
		assert.Nil(t, reader)
		require.EqualError(t, err, "failed to open the file key: internal: failed to open the sealed key")
	})
}

type closer struct {
	isClose bool
}

func (c *closer) Close() error {
	c.isClose = true
	return nil
}

func Test_DecReadSeeker(t *testing.T) {
	content := []byte("Hello, World!")
	closer := closer{false}
	reader := bytes.NewReader(content)
	dec := newDecReadSeeker(reader, int64(len(content)), &closer)

	t.Run("Read", func(t *testing.T) {
		buf := make([]byte, 2)
		n, err := dec.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.Equal(t, []byte("He"), buf)
	})

	t.Run("Seek", func(t *testing.T) {
		n1, err := dec.Seek(4, io.SeekStart)
		require.NoError(t, err)
		assert.Equal(t, int64(4), n1)

		n2, err := dec.Seek(4, io.SeekCurrent)
		require.NoError(t, err)
		assert.Equal(t, int64(8), n2)
	})

	t.Run("Seek and Read", func(t *testing.T) {
		n1, err := dec.Seek(-2, io.SeekEnd)
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)-2), n1)

		buf := make([]byte, 2)
		n2, err := dec.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, 2, n2)
		assert.Equal(t, []byte("d!"), buf)

		t.Run("Read to EOF", func(t *testing.T) {
			buf := make([]byte, 2)
			n2, err := dec.Read(buf)
			assert.Equal(t, err, io.EOF)
			assert.Equal(t, 0, n2)
		})

		t.Run("Seek with an invalid whence", func(t *testing.T) {
			n, err := dec.Seek(2, 4)
			assert.Empty(t, n)
			require.ErrorContains(t, err, "invalid whence")
		})

		t.Run("Seek a negative value", func(t *testing.T) {
			n, err := dec.Seek(-1, io.SeekStart)
			assert.Empty(t, n)
			require.ErrorContains(t, err, "negative position")
		})

		t.Run("Close", func(t *testing.T) {
			err := dec.Close()
			require.NoError(t, err)

			assert.True(t, closer.isClose)
		})
	})
}
