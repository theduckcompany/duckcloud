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
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
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

		inodesMock.On("MkdirAll", mock.Anything, &users.ExampleAlice, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			FilePath:  "foo",
			CreatedBy: &users.ExampleAlice,
		})
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
		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(&files.ExampleFile1, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Space:      spaceFS.space,
			Parent:     &ExampleAliceDir,
			Name:       "bar.txt",
			File:       &files.ExampleFile1,
			UploadedAt: now,
			UploadedBy: &users.ExampleAlice,
		}).Return(&ExampleAliceFile, nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceFile.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			FilePath:   "foo/bar.txt",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
	})

	t.Run("Upload with a validation error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		err := spaceFS.Upload(ctx, &UploadCmd{
			FilePath:   "foo/bar.txt",
			Content:    nil,
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.EqualError(t, err, "validation: Content: cannot be blank.")
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
			MovedBy:     users.ExampleAlice.ID(),
		}).Return(nil).Once()

		err := spaceFS.Move(ctx, &MoveCmd{
			SrcPath: "/foo.txt",
			NewPath: "/bar.txt",
			MovedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
	})

	t.Run("Move with a validation error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		err := spaceFS.Move(ctx, &MoveCmd{
			SrcPath: "",
			NewPath: "/bar.txt",
			MovedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.EqualError(t, err, "validation: SrcPath: cannot be blank.")
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

		err := spaceFS.Move(ctx, &MoveCmd{
			SrcPath: "/foo.txt",
			NewPath: "/bar.txt",
			MovedBy: &users.ExampleAlice,
		})
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
			MovedBy:     users.ExampleAlice.ID(),
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Move(ctx, &MoveCmd{
			SrcPath: "/foo.txt",
			NewPath: "/bar.txt",
			MovedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Rename success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("GetByNameAndParent", mock.Anything, "foobar.jpg", *ExampleAliceFile.Parent()).Return(nil, errs.ErrNotFound).Once()
		inodesMock.On("PatchRename", mock.Anything, &ExampleAliceFile, "foobar.jpg").Return(&ExampleAliceRenamedFile, nil).Once()

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "foobar.jpg")

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceRenamedFile, res)
	})

	t.Run("Rename with an empty name", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "can't be empty")
	})

	t.Run("Rename with a root inode", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		res, err := spaceFS.Rename(ctx, &ExampleAliceRoot, "foo")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "can't rename the root")
	})

	t.Run("Rename with a file with the same name", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		spaceFS := newLocalFS(inodesMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		inodesMock.On("GetByNameAndParent", mock.Anything, "foobar.pdf", *ExampleAliceFile.Parent()).Return(&ExampleAliceFile, nil).Once()
		inodesMock.On("GetByNameAndParent", mock.Anything, "foobar (1).pdf", *ExampleAliceFile.Parent()).Return(nil, errs.ErrNotFound).Once()
		inodesMock.On("PatchRename", mock.Anything, &ExampleAliceFile, "foobar (1).pdf").Return(&ExampleAliceRenamedFile, nil).Once()

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "foobar.pdf")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceRenamedFile, res)
	})
}
