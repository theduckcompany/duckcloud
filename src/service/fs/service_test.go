package fs

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_FS(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)

	userID := uuid.UUID("fd801c11-356a-4abb-8d72-1ea87d2d7201")
	barTxtContent := []byte("Hello, World!")

	inodesSvc := inodes.Init(tools, db)
	afs := afero.NewMemMapFs()
	filesSvc, err := files.NewFSService(afs, "/", tools.Logger())
	require.NoError(t, err)

	rootInode, err := inodesSvc.CreateRootDir(ctx)
	require.NoError(t, err)
	duckFS := NewFSService(userID, rootInode.ID(), inodesSvc, filesSvc)

	var file FileOrDirectory

	t.Run("Stat on root", func(t *testing.T) {
		info, err := duckFS.Stat(ctx, "")
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, "", info.Name())
		assert.Equal(t, int64(0), info.Size(), "size")
		assert.WithinDuration(t, time.Now(), info.ModTime(), 400*time.Millisecond)
	})

	t.Run("Stat on an invalid path", func(t *testing.T) {
		info, err := duckFS.Stat(ctx, "./unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrInvalid)
	})

	t.Run("Stat on an unknown file", func(t *testing.T) {
		// refFS := afero.NewMemMapFs()
		// info, err := refFS.Stat("foo")

		info, err := duckFS.Stat(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		err = duckFS.CreateDir(ctx, "foo", 0o700)
		require.NoError(t, err)
	})

	t.Run("OpenFile success", func(t *testing.T) {
		file, err = duckFS.OpenFile(ctx, "foo/bar.txt", os.O_CREATE|os.O_RDWR, 0o700)
		require.NoError(t, err)
	})

	t.Run("Write success", func(t *testing.T) {
		ret, err := file.Write(barTxtContent)
		require.NoError(t, err)
		require.Equal(t, ret, 13)
	})

	t.Run("Close success", func(t *testing.T) {
		err := file.Close()
		require.NoError(t, err)
	})

	t.Run("fstest", func(t *testing.T) {
		assert.NoError(t, fstest.TestFS(duckFS, "foo/bar.txt"))
	})

	t.Run("ReadFile success", func(t *testing.T) {
		res, err := fs.ReadFile(duckFS, "foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, barTxtContent, res)
	})

	t.Run("ReadFile success", func(t *testing.T) {
		res, err := fs.ReadDir(duckFS, "foo")
		assert.NoError(t, err)
		assert.Len(t, res, 1)

		assert.Equal(t, "bar.txt", res[0].Name())
		assert.False(t, res[0].IsDir())
		assert.True(t, res[0].Type().IsRegular())

		infos, err := res[0].Info()
		require.NoError(t, err)

		assert.WithinDuration(t, time.Now(), infos.ModTime(), 400*time.Millisecond)
		assert.Equal(t, int64(13), infos.Size())
	})

	t.Run("CreateDir with an invalid path", func(t *testing.T) {
		// Base path are invalid
		err := duckFS.CreateDir(ctx, "/foo/bar", 0o700)
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
	})

	t.Run("OpenFile with O_APPEND fail", func(t *testing.T) {
		res, err := duckFS.OpenFile(ctx, "foo/bar.txt", os.O_APPEND|os.O_WRONLY, 0o700)
		assert.Nil(t, res)
		assert.EqualError(t, err, "invalid argument: O_SYNC and O_APPEND not supported")
	})

	t.Run("OpenFile with O_EXCL fail if the file exists", func(t *testing.T) {
		res, err := duckFS.OpenFile(ctx, "foo/bar.txt", os.O_EXCL|os.O_CREATE, 0o700)
		assert.Nil(t, res)
		assert.EqualError(t, err, "open foo/bar.txt: file already exists")
		assert.ErrorIs(t, err, fs.ErrExist)
	})

	t.Run("OpenFile with O_EXCL succeed", func(t *testing.T) {
		res, err := duckFS.OpenFile(ctx, "foo/newbar.txt", os.O_EXCL|os.O_CREATE, 0o700)
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = res.Close()
		assert.NoError(t, err)
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		err := duckFS.RemoveAll(ctx, "foo")
		assert.NoError(t, err)

		//nolint: contextcheck // ??!
		res, err := duckFS.Open("foo/bar.txt")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, fs.ErrNotExist)

		//nolint: contextcheck // ??!
		res, err = duckFS.Open("foo")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("RemoveAll with an invalid path", func(t *testing.T) {
		// Should not start with "./"
		err := duckFS.RemoveAll(ctx, "./foo")
		assert.EqualError(t, err, "open ./foo: invalid argument")
	})
}
