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
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestFileService(t *testing.T) {
	ctx := context.Background()

	t.Run("Upload and Download success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storage, fs, tools, masterkeySvc)

		fileMeta, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		require.NoError(t, err)
		assert.NotNil(t, fileMeta)

		reader, err := svc.Download(ctx, fileMeta)
		require.NoError(t, err)
		res, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)
	})

	t.Run("Upload with a fs error", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storage, fs, tools, masterkeySvc)

		// Create an fs error by removing the write permission
		svc.fs = afero.NewReadOnlyFs(fs)

		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		assert.Empty(t, fileID)
		require.ErrorContains(t, err, "operation not permitted")
		require.ErrorContains(t, err, "internal: failed to create")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storage, fs, tools, masterkeySvc)

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
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storage, fs, tools, masterkeySvc)

		// Create a file
		fileID, err := svc.Upload(ctx, iotest.ErrReader(fmt.Errorf("some-error")))
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "upload error")
		require.ErrorContains(t, err, "some-error")
		assert.Empty(t, fileID)
	})

	t.Run("GetMetadata success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storageMock := NewMockStorage(t)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storageMock, fs, tools, masterkeySvc)

		storageMock.On("GetByID", mock.Anything, ExampleFile1.ID()).Return(&ExampleFile1, nil).Once()

		res, err := svc.GetMetadata(ctx, ExampleFile1.ID())
		require.NoError(t, err)
		assert.Equal(t, &ExampleFile1, res)
	})

	t.Run("GetMetadataByChecksum success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storageMock := NewMockStorage(t)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storageMock, fs, tools, masterkeySvc)

		storageMock.On("GetByChecksum", mock.Anything, "some-checksum").Return(&ExampleFile1, nil).Once()

		res, err := svc.GetMetadataByChecksum(ctx, "some-checksum")
		require.NoError(t, err)
		assert.Equal(t, &ExampleFile1, res)
	})

	t.Run("Download an invalid content", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()
		db := storage.NewTestStorage(t)
		storage := newSqlStorage(db)
		cfgSvc := config.Init(db)
		masterkeySvc, err := masterkey.Init(ctx, cfgSvc, fs, masterkey.Config{DevMode: true}, tools)
		require.NoError(t, err)
		svc := NewFileService(storage, fs, tools, masterkeySvc)

		err = afero.WriteFile(fs, "66/66278d2b-7a4f-4764-ac8a-fc08f224eb66", []byte("not encrypted"), 0o755)
		require.NoError(t, err)

		reader, err := svc.Download(ctx, &ExampleFile2)
		assert.Nil(t, reader)
		require.EqualError(t, err, "failed to open the file key: failed to open the sealed key")
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
