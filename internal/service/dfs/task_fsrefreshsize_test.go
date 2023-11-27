package dfs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

func TestFSRefreshSizeTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Add(time.Minute)

	t.Run("Name", func(t *testing.T) {
		runner := NewFSRefreshSizeTaskRunner(nil, nil, nil)
		assert.Equal(t, "fs-refresh-size", runner.Name())
	})

	t.Run("RunArg success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(inodesMock, filesMock, spacesMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		inodesMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		inodesMock.On("RegisterModification", mock.Anything, &ExampleAliceDir, uint64(42), now).Return(nil).Once()

		// Do the same thing for the parent
		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, *ExampleAliceDir.Parent()).Return(&ExampleAliceRoot, nil).Once()
		inodesMock.On("GetSumChildsSize", mock.Anything, ExampleAliceRoot.ID()).Return(uint64(142), nil).Once()
		inodesMock.On("RegisterModification", mock.Anything, &ExampleAliceRoot, uint64(142), now).Return(nil).Once()

		// ExampleAliceRoot doesnt' have a parent because it's a root node so we stop here.

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			SpaceID:    spaces.ExampleAlicePersonalSpace.ID(),
			INodeID:    ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg with a inode not found", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(inodesMock, filesMock, spacesMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		inodesMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		inodesMock.On("RegisterModification", mock.Anything, &ExampleAliceDir, uint64(42), now).Return(nil).Once()

		// Do the same thing for the parent
		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, *ExampleAliceDir.Parent()).Return(nil, errs.ErrNotFound).Once()
		// The parent have been removed so we stop here.

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			SpaceID:    spaces.ExampleAlicePersonalSpace.ID(),
			INodeID:    ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		assert.NoError(t, err)
	})

	t.Run("RunArg success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(inodesMock, filesMock, spacesMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		inodesMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(0), errors.New("some-error")).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			SpaceID:    spaces.ExampleAlicePersonalSpace.ID(),
			INodeID:    ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with a RegisterModification error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(inodesMock, filesMock, spacesMock)

		spacesMock.On("GetByID", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		inodesMock.On("GetByID", mock.Anything, &spaces.ExampleAlicePersonalSpace, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		inodesMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		inodesMock.On("RegisterModification", mock.Anything, &ExampleAliceDir, uint64(42), now).Return(errors.New("some-error")).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			SpaceID:    spaces.ExampleAlicePersonalSpace.ID(),
			INodeID:    ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		assert.ErrorContains(t, err, "some-error")
	})
}
