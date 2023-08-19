package files

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestFileService(t *testing.T) {
	ctx := context.Background()

	t.Run("Open success", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		svc, err := NewFSService(fs, "/", tools.Logger())
		require.NoError(t, err)

		file, err := svc.Open(ctx, uuid.UUID("735277a1-e94f-423a-b1b9-c69488ad72cc"))
		assert.NoError(t, err)

		file.WriteString("Hello, World!")
		err = file.Close()
		require.NoError(t, err)

		file2, err := svc.Open(ctx, uuid.UUID("735277a1-e94f-423a-b1b9-c69488ad72cc"))
		assert.NoError(t, err)

		buf := make([]byte, 13)
		nb, err := file2.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, nb, 13)
		assert.Equal(t, "Hello, World!", string(buf))
	})

	t.Run("NewFSService setup the dir fanout", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		tools := tools.NewMock(t)

		svc, err := NewFSService(fs, "/", tools.Logger())
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
		svc, err := NewFSService(fs, "/", tools.Logger())
		require.NoError(t, err)
		require.NotNil(t, svc)

		// Second time
		svc, err = NewFSService(fs, "/", tools.Logger())
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
		svc, err := NewFSService(fs, "/invalid/path", tools.Logger())
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /invalid/path: file does not exist")
	})

	t.Run("NewFSService with a file as root path", func(t *testing.T) {
		tools := tools.NewMock(t)
		fs := afero.NewMemMapFs()

		_, err := fs.Create("/foo")
		require.NoError(t, err)

		// First time
		svc, err := NewFSService(fs, "/foo", tools.Logger())
		assert.Nil(t, svc)
		assert.EqualError(t, err, "invalid path: open /foo: it must be a directory")
	})
}
