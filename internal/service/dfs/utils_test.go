package dfs

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner/usercreate"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_Walk(t *testing.T) {
	ctx := context.Background()

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	afs := afero.NewMemMapFs()

	filesSvc, err := files.NewFSService(afs, "/", tools)
	require.NoError(t, err)

	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db)
	schedulerSvc := scheduler.Init(db, tools)
	fsSvc := NewFSService(inodesSvc, filesSvc, foldersSvc, schedulerSvc, tools)
	usersSvc := users.Init(tools, db, schedulerSvc)
	runnerSvc := runner.Init([]runner.TaskRunner{
		usercreate.NewTaskRunner(usersSvc, foldersSvc, inodesSvc),
	}, nil, tools, db)

	user, err := usersSvc.Create(ctx, &users.CreateCmd{
		Username: "Jane-Doe",
		Password: "my little secret",
		IsAdmin:  true,
	})
	require.NoError(t, err)

	// Create all the user stuff
	err = runnerSvc.RunSingleJob(ctx)
	require.NoError(t, err)

	userFolders, err := foldersSvc.GetAllUserFolders(ctx, user.ID(), nil)
	require.NoError(t, err)

	folder := &userFolders[0]

	ffs := fsSvc.GetFolderFS(folder)

	t.Run("with an empty folder", func(t *testing.T) {
		res := []string{}

		err = Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"."}, res)
	})

	t.Run("with a simple file", func(t *testing.T) {
		_, err := inodesSvc.CreateFile(ctx, &inodes.CreateFileCmd{
			Parent:     folder.RootFS(),
			Name:       "foo.txt",
			Size:       0,
			Checksum:   "deadbeef",
			FileID:     uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a"),
			UploadedAt: time.Now(),
		})
		require.NoError(t, err)

		res := []string{}

		err = Walk(ctx, ffs, "foo.txt", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"foo.txt"}, res)
	})

	var dirA *inodes.INode
	t.Run("with an empty directory", func(t *testing.T) {
		dirA, err = ffs.CreateDir(ctx, "dir-a")
		require.NoError(t, err)

		res := []string{}

		err = Walk(ctx, ffs, "dir-a", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"dir-a"}, res)
	})

	t.Run("the root with a file and a dir", func(t *testing.T) {
		res := []string{}

		err = Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{".", "dir-a", "foo.txt"}, res)
	})

	t.Run("do all the sub folders", func(t *testing.T) {
		_, err := inodesSvc.CreateFile(ctx, &inodes.CreateFileCmd{
			Parent:     dirA.ID(),
			Name:       "file-a.txt",
			Size:       0,
			Checksum:   "deadbeef",
			FileID:     uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a"),
			UploadedAt: time.Now(),
		})
		require.NoError(t, err)

		res := []string{}

		err = Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{".", "dir-a", "dir-a/file-a.txt", "foo.txt"}, res)
	})

	t.Run("with a big folder and pagination", func(t *testing.T) {
		dir, err := ffs.CreateDir(ctx, "big-folder")
		require.NoError(t, err)

		for i := 0; i < 100; i++ {
			_, err := inodesSvc.CreateFile(ctx, &inodes.CreateFileCmd{
				Parent:     dir.ID(),
				Name:       fmt.Sprintf("%d.txt", i),
				Size:       0,
				Checksum:   "deadbeef",
				FileID:     uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a"),
				UploadedAt: time.Now(),
			})
			require.NoError(t, err)
		}

		res := []string{}

		err = Walk(ctx, ffs, "big-folder", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Len(t, res, 101) // 100 files + the dir itself
		assert.Contains(t, res, "big-folder")
		for i := 0; i < 100; i++ {
			assert.Contains(t, res, path.Join("big-folder", fmt.Sprintf("%d.txt", i)))
		}
	})
}
