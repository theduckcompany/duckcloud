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
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

func TestFSMoveTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Add(time.Minute)

	t.Run("Name", func(t *testing.T) {
		runner := NewFSMoveTaskRunner(nil, nil, nil, nil)
		assert.Equal(t, "fs-move", runner.Name())
	})

	t.Run("RunArg success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		newFile := inodes.ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/bar.txt",
		}).Return(nil, errs.ErrNotFound).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &users.ExampleAlice, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/",
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
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an existing file at destination", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		newFile := inodes.ExampleAliceFile

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &users.ExampleAlice, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/",
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
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an unknown space", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with an unknown source inode", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with a inodes.Get error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/bar.txt",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with an MkdirAll error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		runner := NewFSMoveTaskRunner(inodesMock, spacesMock, usersMock, schedulerMock)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/bar.txt",
		}).Return(&inodes.ExampleAliceDir, nil).Once()
		inodesMock.On("GetByID", mock.Anything, inodes.ExampleAliceFile.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("MkdirAll", mock.Anything, &users.ExampleAlice, &inodes.PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/",
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: inodes.ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
