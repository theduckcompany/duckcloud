package dfs_test

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
)

func Test_Walk(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)
	serv.Bootstrap(t)

	userFolders, err := serv.FoldersSvc.GetAllUserFolders(ctx, serv.User.ID(), nil)
	require.NoError(t, err)

	folder := &userFolders[0]

	ffs := serv.DFSSvc.GetFolderFS(folder)

	t.Run("with an empty folder", func(t *testing.T) {
		res := []string{}

		err = dfs.Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"."}, res)
	})

	t.Run("with a simple file", func(t *testing.T) {
		err := ffs.Upload(ctx, "/foo.txt", http.NoBody)
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}
		err = dfs.Walk(ctx, ffs, "foo.txt", func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"foo.txt"}, res)
	})

	t.Run("with an empty directory", func(t *testing.T) {
		_, err := ffs.CreateDir(ctx, "dir-a")
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}

		err = dfs.Walk(ctx, ffs, "dir-a", func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"dir-a"}, res)
	})

	t.Run("the root with a file and a dir", func(t *testing.T) {
		res := []string{}

		err = dfs.Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{".", "dir-a", "foo.txt"}, res)
	})

	t.Run("do all the sub folders", func(t *testing.T) {
		err := ffs.Upload(ctx, "/dir-a/file-a.txt", http.NoBody)
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}
		err = dfs.Walk(ctx, ffs, ".", func(_ context.Context, p string, _ *dfs.INode) error {
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
			err := ffs.Upload(ctx, fmt.Sprintf("/big-folder/%d.txt", i), http.NoBody)
			require.NoError(t, err)
		}

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}

		err = dfs.Walk(ctx, ffs, "big-folder", func(_ context.Context, p string, _ *dfs.INode) error {
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
