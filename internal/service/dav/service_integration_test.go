package dav

import (
	"context"
	stdfs "io/fs"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type simpleFS struct {
	fs         *davFS
	davSession *davsessions.DavSession
}

func (s *simpleFS) Open(name string) (stdfs.File, error) {
	ctx := context.WithValue(context.Background(), sessionKeyCtx, s.davSession)
	return s.fs.OpenFile(ctx, name, os.O_RDONLY, 0o644)
}

func Test_DavFS_integration(t *testing.T) {
	ctx := context.Background()

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db, inodesSvc)
	usersSvc := users.Init(tools, db, inodesSvc, foldersSvc)
	davSessionsSvc := davsessions.Init(db, foldersSvc, usersSvc, tools)

	afs := afero.NewMemMapFs()
	filesSvc, err := files.NewFSService(afs, "/", tools.Logger())
	require.NoError(t, err)

	user, err := usersSvc.Create(ctx, &users.CreateCmd{
		Username: "foo-user",
		Password: "my-little-secret",
		IsAdmin:  false,
	})
	require.NoError(t, err)

	folder, err := foldersSvc.CreatePersonalFolder(ctx, &folders.CreatePersonalFolderCmd{
		Name:  "Test",
		Owner: user.ID(),
	})
	require.NoError(t, err)

	session, _, err := davSessionsSvc.Create(ctx, &davsessions.CreateCmd{
		Name:    "test session",
		UserID:  user.ID(),
		Folders: []uuid.UUID{folder.ID()},
	})
	require.NoError(t, err)

	fsSvc := fs.NewFSService(inodesSvc, filesSvc, foldersSvc)
	fs := &davFS{foldersSvc, fsSvc}

	ctx = context.WithValue(ctx, sessionKeyCtx, session)

	t.Run("Create a directory", func(t *testing.T) {
		err := fs.Mkdir(ctx, "foo", 0o644)
		require.NoError(t, err)
	})

	t.Run("Create a file", func(t *testing.T) {
		file, err := fs.OpenFile(ctx, "foo/bar", os.O_CREATE|os.O_WRONLY, 0o644)
		require.NoError(t, err)

		n, err := file.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
		assert.Equal(t, 13, n)
	})

	t.Run("Read a file", func(t *testing.T) {
		file, err := fs.OpenFile(ctx, "foo/bar.txt", os.O_CREATE|os.O_WRONLY, 0o644)
		require.NoError(t, err)

		n, err := file.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
		assert.Equal(t, 13, n)
	})

	t.Run("Readfile with ReadFile", func(t *testing.T) {
		res, err := stdfs.ReadFile(&simpleFS{fs, session}, "foo/bar.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)
	})
}
