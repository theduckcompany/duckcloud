package files

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func TestFileService(t *testing.T) {
	ctx := context.Background()

	t.Run("Upload and Download success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

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

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		// Create an fs error by removing the write permission
		svc.fs = afero.NewReadOnlyFs(fs)

		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		assert.Empty(t, fileID)
		assert.ErrorContains(t, err, "operation not permitted")
		assert.ErrorContains(t, err, "internal: failed to open")
	})

	t.Run("NewFSService setup the dir fanout", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		tools := tools.NewToolboxForTest(t)

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)
		require.NotNil(t, svc)

		dir, err := fs.Open("/")
		require.NoError(t, err)

		res, err := dir.Readdir(300)
		assert.NoError(t, err)
		assert.Len(t, res, 256)
		assert.Equal(t, res[0].Name(), "00")
		assert.Equal(t, res[255].Name(), "ff")
	})

	t.Run("NewFSService can be called with a fs already setup", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()

		// First time
		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)
		require.NotNil(t, svc)

		// Second time
		svc, err = NewFSService(fs, "/", tools)
		require.NoError(t, err)
		require.NotNil(t, svc)

		dir, err := fs.Open("/")
		require.NoError(t, err)

		res, err := dir.Readdir(300)
		assert.NoError(t, err)
		assert.Len(t, res, 256)
		assert.Equal(t, res[0].Name(), "00")
		assert.Equal(t, res[255].Name(), "ff")
	})

	t.Run("NewFSService with an invalid root path", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()

		// First time
		svc, err := NewFSService(fs, "/invalid/path", tools)
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /invalid/path: file does not exist")
	})

	t.Run("NewFSService with a file as root path", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()

		_, err := fs.Create("/foo")
		require.NoError(t, err)

		// First time
		svc, err := NewFSService(fs, "/foo", tools)
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /foo: it must be a directory")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewToolboxForTest(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		// Create a file
		fileID, err := svc.Upload(ctx, bytes.NewReader([]byte("Hello, World!")))
		require.NoError(t, err)
		assert.NotEmpty(t, fileID)

		// Delete it
		err = svc.Delete(ctx, fileID)
		assert.NoError(t, err)
	})
}
