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
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestFSMoveTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Add(time.Minute)

	const AlicePersonalSpaceID = uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec")

	t.Run("Name", func(t *testing.T) {
		runner := NewFSMoveTaskRunner(nil, nil)
		assert.Equal(t, "fs-move", runner.Name())
	})

	t.Run("RunArg success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, schedulerMock)

		newFile := ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			SpaceID: inodes.ExampleAliceDir.SpaceID(),
			Path:    "/bar.txt",
		}).Return(nil, errs.ErrNotFound).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			SpaceID: inodes.ExampleAliceFile.SpaceID(),
			Path:    "/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("PatchMove", mock.Anything, &inodes.ExampleAliceFile, &inodes.ExampleAliceRoot, "bar.txt", now).
			Return(&newFile, nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *inodes.ExampleAliceFile.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      inodes.ExampleAliceRoot.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     inodes.ExampleAliceFile.SpaceID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an existing file at destination", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, schedulerMock)

		newFile := ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()
		inodesMock.On("PatchMove", mock.Anything, &inodes.ExampleAliceFile, &inodes.ExampleAliceRoot, "bar.txt", now).
			Return(&newFile, nil).Once()
		inodesMock.On("Remove", mock.Anything, &inodes.ExampleAliceDir).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *inodes.ExampleAliceFile.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      inodes.ExampleAliceRoot.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     AlicePersonalSpaceID,
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an unknown source inode", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, schedulerMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     AlicePersonalSpaceID,
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with a inodes.Get error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, schedulerMock)

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/bar.txt",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     AlicePersonalSpaceID,
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with an MkdirAll error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, schedulerMock)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &inodes.PathCmd{
			SpaceID: AlicePersonalSpaceID,
			Path:    "/",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     AlicePersonalSpaceID,
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
