package dav

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type simpleFS struct {
	fs         *davFS
	davSession *davsessions.DavSession
}

func (s *simpleFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	ctx := context.WithValue(context.Background(), sessionKeyCtx, s.davSession)
	return s.fs.OpenFile(ctx, name, os.O_RDONLY, 0o644)
}

func Test_DavFS_integration(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)
	serv.Bootstrap(t)

	userFolders, err := serv.FoldersSvc.GetAllUserFolders(ctx, serv.User.ID(), nil)
	require.NoError(t, err)

	session, _, err := serv.DavSessionsSvc.Create(ctx, &davsessions.CreateCmd{
		Name:     "test session",
		UserID:   serv.User.ID(),
		Username: serv.User.Username(),
		Folders:  []uuid.UUID{userFolders[0].ID()},
	})
	require.NoError(t, err)

	davfs := &davFS{serv.FoldersSvc, serv.DFSSvc}

	ctx = context.WithValue(ctx, sessionKeyCtx, session)

	t.Run("Create a directory", func(t *testing.T) {
		err := davfs.Mkdir(ctx, "foo", 0o644)
		require.NoError(t, err)
	})

	t.Run("Write into a file", func(t *testing.T) {
		file, err := davfs.OpenFile(ctx, "foo/bar.1.txt", os.O_CREATE|os.O_WRONLY, 0o644)
		require.NoError(t, err)

		n, err := file.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
		assert.Equal(t, 13, n)

		assert.NoError(t, file.Close())
	})

	t.Run("Running fileupload job make the files available", func(t *testing.T) {
		file, err := davfs.OpenFile(ctx, "foo/bar.2.txt", os.O_CREATE|os.O_WRONLY, 0o644)
		require.NoError(t, err)

		n, err := file.Write([]byte("Hello, World!"))
		require.NoError(t, err)
		require.Equal(t, 13, n)

		require.NoError(t, file.Close())

		//  Run all the pending tasks
		err = serv.RunnerSvc.RunSingleJob(ctx)
		require.NoError(t, err)

		res, err := fs.ReadFile(&simpleFS{davfs, session}, "foo/bar.2.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)
	})

	t.Run("Fstest", func(t *testing.T) {
		err = fstest.TestFS(&simpleFS{davfs, session}, "foo/bar.1.txt", "foo/bar.2.txt")
		if err != nil {
			t.Fatal(err)
		}
	})
}
