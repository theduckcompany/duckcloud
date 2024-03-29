package dfs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestFSGC(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("Name", func(t *testing.T) {
		tools := tools.NewMock(t)
		job := NewFSGGCTaskRunner(nil, nil, nil, nil, tools)
		assert.Equal(t, "fs-gc", job.Name())
	})

	t.Run("Run Success", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{}, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{}`))
		require.NoError(t, err)
	})

	t.Run("Run with some invalid json arg", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{}, nil).Once()

		// It works because we don't need the arg to run the job.
		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		require.NoError(t, err)
	})

	t.Run("RunArgs with a root success", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		storageMock.On("GetAllChildrens", mock.Anything, ExampleAliceRoot.ID(), &sqlstorage.PaginateCmd{Limit: 10}).Return([]INode{ExampleAliceFile}, nil).Once()

		// We remove the file content and inode
		tools.ClockMock.On("Now").Return(now)
		storageMock.On("GetAllInodesWithFileID", mock.Anything, *ExampleAliceFile.FileID()).Return([]INode{}, nil).Once()
		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()
		filesMock.On("Delete", mock.Anything, *ExampleAliceFile.FileID()).Return(nil).Once()

		// We remove the dir itself
		storageMock.On("HardDelete", mock.Anything, ExampleAliceRoot.ID()).Return(nil).Once()

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.NoError(t, err)
	})

	t.Run("RunArgs with a directory success", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceDir}, nil).Once()

		// This is a dir we will delete all its content
		storageMock.On("GetAllChildrens", mock.Anything, ExampleAliceDir.ID(), &sqlstorage.PaginateCmd{Limit: 10}).Return([]INode{ExampleAliceFile}, nil).Once()

		// We remove the file content and inode
		tools.ClockMock.On("Now").Return(now)
		storageMock.On("GetAllInodesWithFileID", mock.Anything, *ExampleAliceFile.FileID()).Return([]INode{}, nil).Once()
		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()
		filesMock.On("Delete", mock.Anything, *ExampleAliceFile.FileID()).Return(nil).Once()

		// We remove the dir itself
		storageMock.On("HardDelete", mock.Anything, ExampleAliceDir.ID()).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.parent,
			ModifiedAt: now,
		}).Return(nil)

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.NoError(t, err)
	})

	t.Run("RunArgs with a file success", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceFile}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		// We remove the file content and inode
		tools.ClockMock.On("Now").Return(now)
		storageMock.On("GetAllInodesWithFileID", mock.Anything, *ExampleAliceFile.FileID()).Return([]INode{}, nil).Once()
		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()
		filesMock.On("Delete", mock.Anything, *ExampleAliceFile.FileID()).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceFile.parent,
			ModifiedAt: now,
		}).Return(nil)

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.NoError(t, err)
	})

	t.Run("with a GetAllDeleted error", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return(nil, fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.EqualError(t, err, "failed to GetAllDeleted: some-error")
	})

	t.Run("with a Readdir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceRoot}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		// This is a dir we will delete all its content
		storageMock.On("GetAllChildrens", mock.Anything, ExampleAliceRoot.ID(), &sqlstorage.PaginateCmd{Limit: 10}).Return(nil, fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.EqualError(t, err, "failed to delete inode \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\": failed to Readdir: some-error")
	})

	t.Run("RunArgs with a RegisterFSRefreshSizeTask error", func(t *testing.T) {
		tools := tools.NewMock(t)
		spacesMock := spaces.NewMockService(t)
		filesMock := files.NewMockService(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFSGGCTaskRunner(storageMock, filesMock, spacesMock, schedulerMock, tools)

		// First loop to fetch the deleted inodes
		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceFile}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		// We remove the file content and inode
		tools.ClockMock.On("Now").Return(now)
		storageMock.On("GetAllInodesWithFileID", mock.Anything, *ExampleAliceFile.FileID()).Return([]INode{}, nil).Once()
		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()
		filesMock.On("Delete", mock.Anything, *ExampleAliceFile.FileID()).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.parent,
			ModifiedAt: now,
		}).Return(fmt.Errorf("some-error"))

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		require.EqualError(t, err, "failed to schedule the fs-refresh-size task: some-error")
	})
}
