package fs

import (
	"bytes"
	"context"
	"io/fs"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func Test_LocalFS(t *testing.T) {
	ctx := context.Background()

	t.Run("Get success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foobar",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := folderFS.Get(ctx, "foobar")
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Get on an invalid path", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		info, err := folderFS.Get(ctx, "./unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrInvalid)
	})

	t.Run("Get on an unknown file", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "unknown-file",
		}).Return(nil, nil).Once()

		info, err := folderFS.Get(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := folderFS.CreateDir(ctx, "foo")
		require.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceRoot, res)
	})

	t.Run("CreateDir with an invalid path", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		// Base path are invalid
		res, err := folderFS.CreateDir(ctx, "/foo/bar")
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
		assert.Nil(t, res)
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("RemoveAll", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}).Return(nil).Once()

		err := folderFS.RemoveAll(ctx, "foo")
		assert.NoError(t, err)
	})

	t.Run("RemoveAll with an invalid path", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		// Should not start with "./"
		err := folderFS.RemoveAll(ctx, "./foo")
		assert.EqualError(t, err, "open ./foo: invalid argument")
	})

	t.Run("ListDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}, &storage.PaginateCmd{Limit: 2}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		res, err := folderFS.ListDir(ctx, "foo", &storage.PaginateCmd{Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, []inodes.INode{inodes.ExampleAliceFile}, res)
	})

	t.Run("Createfile success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Parent: inodes.ExampleAliceRoot.ID(),
			Name:   "bar.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := folderFS.CreateFile(ctx, "foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Createfile with a parent not found", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/",
		}).Return(nil, nil).Once()

		res, err := folderFS.CreateFile(ctx, "foo/bar.txt")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidPath)
		assert.EqualError(t, err, "createFile foo/: invalid path")
	})

	t.Run("Download success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		file, err := afero.TempFile(afero.NewMemMapFs(), "foo", "")
		require.NoError(t, err)

		filesMock.On("Open", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(file, nil).Once()

		res, err := folderFS.Download(ctx, &inodes.ExampleAliceFile)
		assert.NoError(t, err)
		assert.Equal(t, file, res)
	})

	t.Run("Upload success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock)

		content := "Hello, World!"

		fs := afero.NewMemMapFs()
		file, err := afero.TempFile(fs, "foo", "")
		require.NoError(t, err)

		filesMock.On("Open", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(file, nil).Once()

		inodesMock.On("RegisterWrite", mock.Anything, &inodes.ExampleAliceFile, int64(len(content)), mock.Anything).Return(nil).Once()
		foldersMock.On("RegisterWrite", mock.Anything, folders.ExampleAlicePersonalFolder.ID(), uint64(len(content))).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		err = folderFS.Upload(ctx, &inodes.ExampleAliceFile, bytes.NewBufferString(content))
		assert.NoError(t, err)
	})
}
