package dfs

import (
	"bytes"
	"context"
	"io/fs"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_LocalFS(t *testing.T) {
	ctx := context.Background()

	t.Run("Get success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

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
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		info, err := folderFS.Get(ctx, "./unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, fs.ErrInvalid)
	})

	t.Run("Get on an unknown file", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "unknown-file",
		}).Return(nil, errs.ErrNotFound).Once()

		info, err := folderFS.Get(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

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
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		// Base path are invalid
		res, err := folderFS.CreateDir(ctx, "/foo/bar")
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
		assert.Nil(t, res)
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

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
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		// Should not start with "./"
		err := folderFS.RemoveAll(ctx, "./foo")
		assert.EqualError(t, err, "open ./foo: invalid argument")
	})

	t.Run("ListDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}, &storage.PaginateCmd{Limit: 2}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		res, err := folderFS.ListDir(ctx, "foo", &storage.PaginateCmd{Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, []inodes.INode{inodes.ExampleAliceFile}, res)
	})

	t.Run("Download success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		file, err := afero.TempFile(afero.NewMemMapFs(), "foo", "")
		require.NoError(t, err)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folderFS.folder.RootFS(),
			FullName: "/foo/bar.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		filesMock.On("Open", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return(file, nil).Once()

		res, err := folderFS.Download(ctx, "/foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, file, res)
	})

	t.Run("Upload success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, uploadsMock)

		content := "Hello, World!"

		fs := afero.NewMemMapFs()
		file, err := afero.TempFile(fs, "foo", "")
		require.NoError(t, err)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		filesMock.On("Create", mock.Anything).Return(file, uuid.UUID("some-file-id"), nil).Once()

		uploadsMock.On("Register", mock.Anything, &uploads.RegisterUploadCmd{
			FolderID: folders.ExampleAlicePersonalFolder.ID(),
			DirID:    inodes.ExampleAliceRoot.ID(),
			FileName: "bar.txt",
			FileID:   uuid.UUID("some-file-id"),
		}).Return(nil).Once()

		err = folderFS.Upload(ctx, "foo/bar.txt", bytes.NewBufferString(content))
		assert.NoError(t, err)
	})
}
