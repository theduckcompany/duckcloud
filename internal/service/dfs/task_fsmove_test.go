package dfs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

func TestFSMoveTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Add(time.Minute)

	t.Run("Name", func(t *testing.T) {
		runner := NewFSMoveTaskRunner(nil, nil)
		assert.Equal(t, "fs-move", runner.Name())
	})

	t.Run("RunArg success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		newFile := ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/bar.txt",
		}).Return(nil, errs.ErrNotFound).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("RegisterDeletion", mock.Anything, &inodes.ExampleAliceFile, uint64(42), now).
			Return(nil).Once()
		inodesMock.On("PatchMove", mock.Anything, &inodes.ExampleAliceFile, &inodes.ExampleAliceRoot, "bar.txt", now).
			Return(&newFile, nil).Once()
		inodesMock.On("RegisterWrite", mock.Anything, &newFile, uint64(42), now.Add(time.Microsecond)).
			Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an existing file at destination", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		newFile := ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("RegisterDeletion", mock.Anything, &inodes.ExampleAliceFile, uint64(42), now).
			Return(nil).Once()
		inodesMock.On("PatchMove", mock.Anything, &inodes.ExampleAliceFile, &inodes.ExampleAliceRoot, "bar.txt", now).
			Return(&newFile, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceDir).Return(nil).Once()
		inodesMock.On("RegisterWrite", mock.Anything, &newFile, uint64(42), now.Add(time.Microsecond)).
			Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an unknown folder", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with an unknown source inode", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with a inodes.Get error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/bar.txt",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with an MkdirAll error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		runner := NewFSMoveTaskRunner(inodesMock, foldersMock)

		foldersMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			Folder: &folders.ExampleAlicePersonalFolder,
			Path:   "/",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			FolderID:    folders.ExampleAlicePersonalFolder.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
