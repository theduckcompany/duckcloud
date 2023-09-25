package fs

import (
	"context"
	"io"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

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

type simplefs struct {
	fs FS
}

func (f *simplefs) Open(name string) (fs.File, error) {
	return f.fs.OpenFile(context.Background(), name, os.O_RDONLY)
}

func Test_FS(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)

	userID := uuid.UUID("fd801c11-356a-4abb-8d72-1ea87d2d7201")
	barTxtContent := []byte("Hello, World!")

	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db, inodesSvc)

	afs := afero.NewMemMapFs()
	filesSvc, err := files.NewFSService(afs, "/", tools.Logger())
	require.NoError(t, err)

	folder, err := foldersSvc.CreatePersonalFolder(ctx, &folders.CreatePersonalFolderCmd{
		Name:  "My folder",
		Owner: userID,
	})
	require.NoError(t, err)

	duckFS := NewFSService(inodesSvc, filesSvc, foldersSvc)

	var file FileOrDirectory

	folderFS := duckFS.GetFolderFS(folder)

	t.Run("Stat on root", func(t *testing.T) {
		info, err := folderFS.Stat(ctx, "")
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, "", info.Name())
		assert.Equal(t, int64(0), info.Size(), "size")
		assert.WithinDuration(t, time.Now(), info.ModTime(), 400*time.Millisecond)
	})

	t.Run("Stat on an invalid path", func(t *testing.T) {
		info, err := folderFS.Stat(ctx, "./unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrInvalid)
	})

	t.Run("Stat on an unknown file", func(t *testing.T) {
		info, err := folderFS.Stat(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		res, err := folderFS.CreateDir(ctx, "foo")
		require.NoError(t, err)
		assert.Equal(t, "foo", res.Name())
	})

	t.Run("OpenFile success", func(t *testing.T) {
		file, err = folderFS.OpenFile(ctx, "foo/bar.txt", os.O_CREATE|os.O_RDWR)
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
		assert.NoError(t, fstest.TestFS(&simplefs{folderFS}, "foo/bar.txt"))
	})

	t.Run("ReadFile success", func(t *testing.T) {
		res, err := fs.ReadFile(&simplefs{folderFS}, "foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, barTxtContent, res)
	})

	t.Run("ReadFile success", func(t *testing.T) {
		res, err := fs.ReadDir(&simplefs{folderFS}, "foo")
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
		res, err := folderFS.CreateDir(ctx, "/foo/bar")
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
		assert.Nil(t, res)
	})

	t.Run("OpenFile with O_APPEND fail", func(t *testing.T) {
		res, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_APPEND|os.O_WRONLY)
		assert.Nil(t, res)
		assert.EqualError(t, err, "invalid argument: O_SYNC and O_APPEND not supported")
	})

	t.Run("OpenFile with O_EXCL fail if the file exists", func(t *testing.T) {
		res, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_EXCL|os.O_CREATE)
		assert.Nil(t, res)
		assert.EqualError(t, err, "open foo/bar.txt: file already exists")
		assert.ErrorIs(t, err, fs.ErrExist)
	})

	t.Run("OpenFile with O_EXCL succeed", func(t *testing.T) {
		res, err := folderFS.OpenFile(ctx, "foo/newbar.txt", os.O_EXCL|os.O_CREATE)
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = res.Close()
		assert.NoError(t, err)
	})

	t.Run("OpenFile with O_TRUNC succeed", func(t *testing.T) {
		// Check the state first
		startfile, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_RDONLY)
		require.NoError(t, err)
		startRes, err := io.ReadAll(startfile)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(startRes))
		require.NoError(t, file.Close())

		// Then truncate an put some new content
		file, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_TRUNC)
		require.NoError(t, err)
		file.Write([]byte("foobar"))
		require.NoError(t, file.Close())

		// Then read the new content
		res, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_RDONLY)
		require.NoError(t, err)
		rawRes, err := io.ReadAll(res)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		assert.Equal(t, "foobar", string(rawRes))
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		err := folderFS.RemoveAll(ctx, "foo")
		assert.NoError(t, err)

		res, err := folderFS.OpenFile(ctx, "foo/bar.txt", os.O_RDONLY)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, fs.ErrNotExist)

		res, err = folderFS.OpenFile(ctx, "foo", os.O_RDONLY)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("RemoveAll with an invalid path", func(t *testing.T) {
		// Should not start with "./"
		err := folderFS.RemoveAll(ctx, "./foo")
		assert.EqualError(t, err, "open ./foo: invalid argument")
	})
}
