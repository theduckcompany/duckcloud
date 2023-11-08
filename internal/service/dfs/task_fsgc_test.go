package dfs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestFSGC(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("Name", func(t *testing.T) {
		tools := tools.NewMock(t)
		job := NewFSGGCTaskRunner(nil, nil, nil, tools)
		assert.Equal(t, "fs-gc", job.Name())
	})

	t.Run("Run Success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		job := NewFSGGCTaskRunner(inodesMock, filesMock, foldersMock, tools)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{}, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{}`))
		assert.NoError(t, err)
	})

	t.Run("Run with some invalid json arg", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		job := NewFSGGCTaskRunner(inodesMock, filesMock, foldersMock, tools)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{}, nil).Once()

		// It works because we don't need the arg to run the job.
		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.NoError(t, err)
	})

	t.Run("RunArgs Success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		job := NewFSGGCTaskRunner(inodesMock, filesMock, foldersMock, tools)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesMock.On("Readdir", mock.Anything, &inodes.ExampleAliceRoot, &storage.PaginateCmd{Limit: 10}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		// We remove the file content and inode
		tools.ClockMock.On("Now").Return(now)
		inodesMock.On("HardDelete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()
		inodesMock.On("RegisterDeletion", mock.Anything, &inodes.ExampleAliceFile, inodes.ExampleAliceFile.Size(), now).Return(nil).Once()
		filesMock.On("Delete", mock.Anything, *inodes.ExampleAliceFile.FileID()).Return(nil).Once()

		// We remove the dir itself
		inodesMock.On("HardDelete", mock.Anything, inodes.ExampleAliceRoot.ID()).Return(nil).Once()

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		assert.NoError(t, err)
	})

	t.Run("with a GetAllDeleted error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return(nil, fmt.Errorf("some-error")).Once()

		job := NewFSGGCTaskRunner(inodesMock, filesMock, foldersMock, tools)

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		assert.EqualError(t, err, "failed to GetAllDeleted: some-error")
	})

	t.Run("with a Readdir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesMock.On("Readdir", mock.Anything, &inodes.ExampleAliceRoot, &storage.PaginateCmd{Limit: 10}).Return(nil, fmt.Errorf("some-error")).Once()

		job := NewFSGGCTaskRunner(inodesMock, filesMock, foldersMock, tools)

		err := job.RunArgs(ctx, &scheduler.FSGCArgs{})
		assert.EqualError(t, err, "failed to delete inode \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\": failed to Readdir: some-error")
	})
}
