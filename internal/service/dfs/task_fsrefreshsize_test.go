package dfs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/stats"
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
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceDir.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(nil).Once()

		// Do the same thing for the parent
		storageMock.On("GetByID", mock.Anything, *ExampleAliceDir.Parent()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceRoot.ID()).Return(uint64(142), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(142),
		}).Return(nil).Once()

		// ExampleAliceRoot doesnt' have a parent because it's a root node so we stop here.

		storageMock.On("GetSumRootsSize", mock.Anything).Return(uint64(142), nil).Once()
		statsMock.On("SetTotalSize", mock.Anything, uint64(142)).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.NoError(t, err)
	})

	t.Run("RunArg with an inode not found", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceDir.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(nil).Once()

		// Do the same thing for the parent
		storageMock.On("GetByID", mock.Anything, *ExampleAliceDir.Parent()).Return(nil, errs.ErrNotFound).Once()
		// The parent have been removed so we stop here.

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.NoError(t, err)
	})

	t.Run("RunArg with a GetSumChildsSize error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(0), errors.New("some-error")).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with a RegisterModification error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceDir.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArg with a GetSumRootsSize error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceDir.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(nil).Once()

		// Do the same thing for the parent
		storageMock.On("GetByID", mock.Anything, *ExampleAliceDir.Parent()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceRoot.ID()).Return(uint64(142), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(142),
		}).Return(nil).Once()

		// ExampleAliceRoot doesnt' have a parent because it's a root node so we stop here.

		storageMock.On("GetSumRootsSize", mock.Anything).Return(uint64(0), errs.ErrInternal).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("RunArg with a SetTotalSize error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		storageMock := NewMockStorage(t)
		statsMock := stats.NewMockService(t)
		runner := NewFSRefreshSizeTaskRunner(storageMock, filesMock, statsMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceDir.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceDir.ID()).Return(uint64(42), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceDir.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(nil).Once()

		// Do the same thing for the parent
		storageMock.On("GetByID", mock.Anything, *ExampleAliceDir.Parent()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetSumChildsSize", mock.Anything, ExampleAliceRoot.ID()).Return(uint64(142), nil).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(142),
		}).Return(nil).Once()

		// ExampleAliceRoot doesnt' have a parent because it's a root node so we stop here.

		storageMock.On("GetSumRootsSize", mock.Anything).Return(uint64(142), nil).Once()
		statsMock.On("SetTotalSize", mock.Anything, uint64(142)).Return(errs.ErrBadRequest).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceDir.ID(),
			ModifiedAt: now,
		})
		require.ErrorIs(t, err, errs.ErrBadRequest)
	})
}
