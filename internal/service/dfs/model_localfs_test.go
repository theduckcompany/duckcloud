package dfs

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_LocalFS(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("Get success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foobar",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := folderFS.Get(ctx, "foobar")
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Get on an unknown file", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/unknown-file",
		}).Return(nil, errs.ErrNotFound).Once()

		info, err := folderFS.Get(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := folderFS.CreateDir(ctx, "foo")
		require.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceRoot, res)
	})

	t.Run("Remove success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo",
		}).Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceFile).Return(nil).Once()

		err := folderFS.Remove(ctx, "foo")
		assert.NoError(t, err)
	})

	t.Run("ListDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

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
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

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
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		fs := afero.NewMemMapFs()
		file, err := afero.TempFile(fs, "foo", "")
		require.NoError(t, err)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		filesMock.On("Create", mock.Anything).Return(file, uuid.UUID("some-file-id"), nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()

		schedulerMock.On("RegisterFileUploadTask", mock.Anything, &scheduler.FileUploadArgs{
			FolderID:   folders.ExampleAlicePersonalFolder.ID(),
			Directory:  inodes.ExampleAliceRoot.ID(),
			FileName:   "bar.txt",
			FileID:     uuid.UUID("some-file-id"),
			UploadedAt: now,
		}).Return(nil).Once()

		err = folderFS.Upload(ctx, "foo/bar.txt", bytes.NewBufferString(content))
		assert.NoError(t, err)
	})

	t.Run("Rename success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		inodesMock.On("Move", mock.Anything, &inodes.ExampleAliceFile, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/bar.txt",
		}).Return(nil).Once()

		err := folderFS.Rename(ctx, "/foo.txt", "/bar.txt")
		assert.NoError(t, err)
	})

	t.Run("Rename with a source not found", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo.txt",
		}).Return(nil, errs.ErrNotFound).Once()

		err := folderFS.Rename(ctx, "/foo.txt", "/bar.txt")
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("Rename with a move error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		folderFS := newLocalFS(inodesMock, filesMock, &folders.ExampleAlicePersonalFolder, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/foo.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		inodesMock.On("Move", mock.Anything, &inodes.ExampleAliceFile, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "/bar.txt",
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := folderFS.Rename(ctx, "/foo.txt", "/bar.txt")
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
