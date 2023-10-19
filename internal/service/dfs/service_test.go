package dfs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/files"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDFSService(t *testing.T) {
	ctx := context.Background()

	// Copy the id to avoid a dependency cycle
	AliceUserID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	t.Run("CreateFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("CreateRootDir", mock.Anything).Return(&inodes.ExampleAliceRoot, nil).Once()
		foldersMock.On("Create", mock.Anything, &folders.CreateCmd{
			Name:   DefaultFolderName,
			Owners: []uuid.UUID{AliceUserID},
			RootFS: inodes.ExampleAliceRoot.ID(),
		}).Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		res, err := svc.CreateFS(ctx, []uuid.UUID{AliceUserID})
		assert.NoError(t, err)
		assert.Equal(t, &folders.ExampleAlicePersonalFolder, res)
	})

	t.Run("CreateFS with a create RootDirError", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("CreateRootDir", mock.Anything).Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		res, err := svc.CreateFS(ctx, []uuid.UUID{AliceUserID})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("CreateFS with a folder create error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("CreateRootDir", mock.Anything).Return(&inodes.ExampleAliceRoot, nil).Once()
		foldersMock.On("Create", mock.Anything, &folders.CreateCmd{
			Name:   DefaultFolderName,
			Owners: []uuid.UUID{AliceUserID},
			RootFS: inodes.ExampleAliceRoot.ID(),
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceRoot).Return(nil).Once()

		res, err := svc.CreateFS(ctx, []uuid.UUID{AliceUserID})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.RootFS()).
			Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceRoot).Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		err := svc.RemoveFS(ctx, &folders.ExampleAlicePersonalFolder)
		assert.NoError(t, err)
	})

	t.Run("RemoveFS with a rootfs not found", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.RootFS()).
			Return(nil, errs.ErrNotFound).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		err := svc.RemoveFS(ctx, &folders.ExampleAlicePersonalFolder)
		assert.NoError(t, err)
	})

	t.Run("RemoveFS with a GetByID error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.RootFS()).
			Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := svc.RemoveFS(ctx, &folders.ExampleAlicePersonalFolder)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS with an GetByID", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.RootFS()).
			Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceRoot).
			Return(errs.Internal(errors.New("some-error"))).Once()

		err := svc.RemoveFS(ctx, &folders.ExampleAlicePersonalFolder)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS with an GetByID", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, foldersMock, schedulerMock, toolsMock)

		inodesMock.On("GetByID", mock.Anything, folders.ExampleAlicePersonalFolder.RootFS()).
			Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceRoot).Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).
			Return(errs.Internal(errors.New("some-error"))).Once()

		err := svc.RemoveFS(ctx, &folders.ExampleAlicePersonalFolder)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
