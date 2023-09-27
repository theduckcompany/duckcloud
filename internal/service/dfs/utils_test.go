package dfs

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_Walk(t *testing.T) {
	ctx := context.Background()

	userID := uuid.UUID("02bd2941-ef0f-4e75-af3e-8b9f835c2ea6")

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db, inodesSvc)

	afs := afero.NewMemMapFs()
	filesSvc, err := files.NewFSService(afs, "/", tools.Logger())
	require.NoError(t, err)

	folder, err := foldersSvc.CreatePersonalFolder(ctx, &folders.CreatePersonalFolderCmd{
		Name:  "Test",
		Owner: userID,
	})
	require.NoError(t, err)

	fsSvc := NewFSService(inodesSvc, filesSvc, foldersSvc)

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
		_, err := ffs.CreateFile(ctx, "foo.txt")
		require.NoError(t, err)

		res := []string{}

		err = Walk(ctx, ffs, "foo.txt", func(_ context.Context, p string, _ *inodes.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"foo.txt"}, res)
	})

	t.Run("with an empty directory", func(t *testing.T) {
		_, err := ffs.CreateDir(ctx, "dir-a")
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
		_, err := ffs.CreateFile(ctx, "dir-a/file-a.txt")
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
		_, err := ffs.CreateDir(ctx, "big-folder")
		require.NoError(t, err)

		for i := 0; i < 100; i++ {
			_, err := ffs.CreateFile(ctx, path.Join("big-folder", fmt.Sprintf("%d.txt", i)))
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
