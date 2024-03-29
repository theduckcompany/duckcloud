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

	userSpaces, err := serv.SpacesSvc.GetAllUserSpaces(ctx, serv.User.ID(), nil)
	require.NoError(t, err)

	space := &userSpaces[0]

	fsService := serv.DFSSvc

	t.Run("with an empty space", func(t *testing.T) {
		res := []string{}

		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "."), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"/"}, res)
	})

	t.Run("with a simple file", func(t *testing.T) {
		err := fsService.Upload(ctx, &dfs.UploadCmd{
			Path:       dfs.NewPathCmd(space, "/foo.txt"),
			Content:    http.NoBody,
			UploadedBy: serv.User,
		})
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}
		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "foo.txt"), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"/foo.txt"}, res)
	})

	t.Run("with an empty directory", func(t *testing.T) {
		_, err := fsService.CreateDir(ctx, &dfs.CreateDirCmd{Path: dfs.NewPathCmd(space, "dir-a"), CreatedBy: serv.User})
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}

		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "dir-a"), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"/dir-a"}, res)
	})

	t.Run("the root with a file and a dir", func(t *testing.T) {
		res := []string{}

		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "."), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"/", "/dir-a", "/foo.txt"}, res)
	})

	t.Run("do all the sub spaces", func(t *testing.T) {
		err := fsService.Upload(ctx, &dfs.UploadCmd{
			Path:       dfs.NewPathCmd(space, "/dir-a/file-a.txt"),
			Content:    http.NoBody,
			UploadedBy: serv.User,
		})
		require.NoError(t, err)

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}
		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "."), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"/", "/dir-a", "/dir-a/file-a.txt", "/foo.txt"}, res)
	})

	t.Run("with a big space and pagination", func(t *testing.T) {
		_, err := fsService.CreateDir(ctx, &dfs.CreateDirCmd{
			Path:      dfs.NewPathCmd(space, "big-space"),
			CreatedBy: serv.User,
		})
		require.NoError(t, err)

		for i := 0; i < 100; i++ {
			err := fsService.Upload(ctx, &dfs.UploadCmd{
				Path:       dfs.NewPathCmd(space, fmt.Sprintf("/big-space/%d.txt", i)),
				Content:    http.NoBody,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)
		}

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)

		res := []string{}

		err = dfs.Walk(ctx, fsService, dfs.NewPathCmd(space, "big-space"), func(_ context.Context, p string, _ *dfs.INode) error {
			res = append(res, p)
			return nil
		})

		require.NoError(t, err)
		assert.Len(t, res, 101) // 100 files + the dir itself
		assert.Contains(t, res, "/big-space")
		for i := 0; i < 100; i++ {
			assert.Contains(t, res, path.Join("/big-space", fmt.Sprintf("%d.txt", i)))
		}
	})
}
