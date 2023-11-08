package files

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestFileService(t *testing.T) {
	const someFileID = uuid.UUID("fa603efe-d91b-4530-baaa-820c297214bd")

	ctx := context.Background()

	t.Run("Open success", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		file, err := svc.Open(ctx, someFileID)
		assert.NoError(t, err)

		file.WriteString("Hello, World!")
		err = file.Close()
		require.NoError(t, err)

		file2, err := svc.Open(ctx, someFileID)
		assert.NoError(t, err)

		buf := make([]byte, 13)
		nb, err := file2.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, nb, 13)
		assert.Equal(t, "Hello, World!", string(buf))
	})

	t.Run("Open with a fs error", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		// Create an fs error by removing the write permission
		svc.fs = afero.NewReadOnlyFs(fs)

		file, err := svc.Open(ctx, uuid.UUID("0367b0e5-4566-449b-ba4c-260010635f01"))
		assert.Nil(t, file)
		assert.EqualError(t, err, "internal: failed to Open \"03/0367b0e5-4566-449b-ba4c-260010635f01\": operation not permitted")
	})

	t.Run("NewFSService setup the dir fanout", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		tools := tools.NewMock(t)

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
		tools := tools.NewMock(t)
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
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		// First time
		svc, err := NewFSService(fs, "/invalid/path", tools)
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /invalid/path: file does not exist")
	})

	t.Run("NewFSService with a file as root path", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		_, err := fs.Create("/foo")
		require.NoError(t, err)

		// First time
		svc, err := NewFSService(fs, "/foo", tools)
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /foo: it must be a directory")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		// Create a file
		file, err := svc.Open(ctx, someFileID)
		assert.NoError(t, err)
		file.WriteString("Hello, World!")
		err = file.Close()
		require.NoError(t, err)

		// Delete it
		err = svc.Delete(ctx, someFileID)
		assert.NoError(t, err)
	})

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools)
		require.NoError(t, err)

		tools.UUIDMock.On("New").Return(uuid.UUID("60d04893-f015-4c09-b68c-6841a08643f3")).Once()

		file, resUUID, err := svc.Create(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Equal(t, uuid.UUID("60d04893-f015-4c09-b68c-6841a08643f3"), resUUID)

		_, err = file.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
		require.NoError(t, file.Close())

		res, err := afero.ReadFile(fs, "/60/60d04893-f015-4c09-b68c-6841a08643f3")
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(res))
	})
}
