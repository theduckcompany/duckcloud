package dav

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner/fileupload"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
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

func (s *simpleFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	ctx := context.WithValue(context.Background(), sessionKeyCtx, s.davSession)
	return s.fs.OpenFile(ctx, name, os.O_RDONLY, 0o644)
}

func Test_DavFS_integration(t *testing.T) {
	ctx := context.Background()

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	afs := afero.NewMemMapFs()
	db := storage.NewTestStorage(t)

	filesSvc, err := files.NewFSService(afs, "/", tools)
	require.NoError(t, err)
	schedulerSvc := scheduler.Init(db, tools)
	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db, inodesSvc)
	usersSvc := users.Init(tools, db, schedulerSvc)
	davSessionsSvc := davsessions.Init(db, foldersSvc, usersSvc, tools)
	runnerSvc := runner.Init([]runner.TaskRunner{fileupload.NewTaskRunner(foldersSvc, filesSvc, inodesSvc)}, tools, db)

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

	fsSvc := dfs.NewFSService(inodesSvc, filesSvc, foldersSvc, schedulerSvc, tools)
	davfs := &davFS{foldersSvc, fsSvc}

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
		err = runnerSvc.RunSingleJob(ctx)
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
