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
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
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
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foobar",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := spaceFS.Get(ctx, "foobar")
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Get on an unknown file", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/unknown-file",
		}).Return(nil, errs.ErrNotFound).Once()

		info, err := spaceFS.Get(ctx, "unknown-file")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := spaceFS.CreateDir(ctx, "foo")
		require.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceRoot, res)
	})

	t.Run("Remove success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo",
		}).Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceFile).Return(nil).Once()

		err := spaceFS.Remove(ctx, "foo")
		assert.NoError(t, err)
	})

	t.Run("ListDir success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("Readdir", mock.Anything, &inodes.ExampleAliceDir, &storage.PaginateCmd{Limit: 2}).
			Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		res, err := spaceFS.ListDir(ctx, "foo", &storage.PaginateCmd{Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, []inodes.INode{inodes.ExampleAliceFile}, res)
	})

	t.Run("Download success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		file, err := afero.TempFile(afero.NewMemMapFs(), "foo", "")
		require.NoError(t, err)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: spaceFS.space,
			Path:  "/foo/bar.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, *inodes.ExampleAliceFile.FileID()).Return(&files.ExampleFile1, nil).Once()

		filesMock.On("Download", mock.Anything, &files.ExampleFile1).Return(file, nil).Once()

		res, err := spaceFS.Download(ctx, "/foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, file, res)
	})

	t.Run("Upload success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/",
		}).Return(&ExampleAliceDir, nil).Once()
		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(uuid.UUID("some-file-id"), nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Space:      spaceFS.space,
			Parent:     ExampleAliceDir.ID(),
			Name:       "bar.txt",
			FileID:     uuid.UUID("some-file-id"),
			UploadedAt: now,
		}).Return(&ExampleAliceFile, nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceFile.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := spaceFS.Upload(ctx, "foo/bar.txt", bytes.NewBufferString(content))
		assert.NoError(t, err)
	})

	t.Run("Move success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		schedulerMock.On("RegisterFSMoveTask", mock.Anything, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		}).Return(nil).Once()

		err := spaceFS.Move(ctx, "/foo.txt", "/bar.txt")
		assert.NoError(t, err)
	})

	t.Run("Move with a source not found", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo.txt",
		}).Return(nil, errs.ErrNotFound).Once()

		err := spaceFS.Move(ctx, "/foo.txt", "/bar.txt")
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("Move with a move error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo.txt",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		schedulerMock.On("RegisterFSMoveTask", mock.Anything, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Move(ctx, "/foo.txt", "/bar.txt")
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
