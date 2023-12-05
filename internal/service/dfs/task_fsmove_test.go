package dfs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
)

func TestFSMoveTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Add(time.Minute)

	t.Run("Name", func(t *testing.T) {
		runner := NewFSMoveTaskRunner(nil, nil, nil, nil, nil)
		assert.Equal(t, "fs-move", runner.Name())
	})

	t.Run("RunArg success", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		fsMock.On("Get", mock.Anything, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"}).Return(nil, errs.ErrNotFound).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceFile.ID()).
			Return(&ExampleAliceFile, nil).Once()
		fsMock.On("CreateDir", mock.Anything, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/",
			CreatedBy: &users.ExampleAlice,
		}).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"parent":           ptr.To(ExampleAliceRoot.ID()),
			"name":             "bar.txt",
			"last_modified_at": now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceFile.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceRoot.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an existing file at destination", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		fsMock.On("Get", mock.Anything, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"}).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceFile.ID()).
			Return(&ExampleAliceFile, nil).Once()
		fsMock.On("CreateDir", mock.Anything, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/",
			CreatedBy: &users.ExampleAlice,
		}).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"parent":           ptr.To(ExampleAliceRoot.ID()),
			"name":             "bar.txt",
			"last_modified_at": now,
		}).Return(nil).Once()
		fsMock.On("removeINode", mock.Anything, &ExampleAliceDir).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceFile.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceRoot.ID(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with an unknown space", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with an unknown source inode", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		fsMock.On("Get", mock.Anything, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"}).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceFile.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArg with a inodes.Get error", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		fsMock.On("Get", mock.Anything, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with an CreateDir error", func(t *testing.T) {
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		usersMock := users.NewMockService(t)
		fsMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		runner := NewFSMoveTaskRunner(fsMock, storageMock, spacesMock, usersMock, schedulerMock)

		require.True(t, ExampleAliceRoot.LastModifiedAt().Before(now))

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		fsMock.On("Get", mock.Anything, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"}).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceFile.ID()).
			Return(&ExampleAliceFile, nil).Once()
		fsMock.On("CreateDir", mock.Anything, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/",
			CreatedBy: &users.ExampleAlice,
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
