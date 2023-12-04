package dfs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

func Test_FSRemoveDuplicateFilesRunner_Task(t *testing.T) {
	ctx := context.Background()

	t.Run("name", func(t *testing.T) {
		runner := NewFSRemoveDuplicateFileRunner(nil, nil, nil)

		assert.Equal(t, "fs-remove-duplicate-file", runner.Name())
	})

	t.Run("RunArgs success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		inodesMock.On("GetAllInodesWithFileID", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, files.ExampleFile1.ID()).
			Return(&files.ExampleFile1, nil).Once()

		inodesMock.On("PatchFileID", mock.Anything, &inodes.ExampleAliceFile, files.ExampleFile1.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()

		filesMock.On("Delete", mock.Anything, *inodes.ExampleAliceFile.FileID()).Return(nil).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRemoveDuplicateFileArgs{
			DuplicateFileID: *inodes.ExampleAliceFile.FileID(),
			ExistingFileID:  files.ExampleFile1.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("RunArgs with a GetAllInodesWithFileID error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		inodesMock.On("GetAllInodesWithFileID", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRemoveDuplicateFileArgs{
			DuplicateFileID: *inodes.ExampleAliceFile.FileID(),
			ExistingFileID:  files.ExampleFile1.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArgs with a GetMetadata error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		inodesMock.On("GetAllInodesWithFileID", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, files.ExampleFile1.ID()).
			Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRemoveDuplicateFileArgs{
			DuplicateFileID: *inodes.ExampleAliceFile.FileID(),
			ExistingFileID:  files.ExampleFile1.ID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RunArgs with a PatchFileID error", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		inodesMock.On("GetAllInodesWithFileID", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, files.ExampleFile1.ID()).
			Return(&files.ExampleFile1, nil).Once()

		inodesMock.On("PatchFileID", mock.Anything, &inodes.ExampleAliceFile, files.ExampleFile1.ID()).
			Return(nil, errs.Internal(errors.New("some-error"))).Once()

		err := runner.RunArgs(ctx, &scheduler.FSRemoveDuplicateFileArgs{
			ExistingFileID:  files.ExampleFile1.ID(),
			DuplicateFileID: *inodes.ExampleAliceFile.FileID(),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Run success", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		inodesMock.On("GetAllInodesWithFileID", mock.Anything, *inodes.ExampleAliceFile.FileID()).
			Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, files.ExampleFile2.ID()).
			Return(&files.ExampleFile1, nil).Once()

		inodesMock.On("PatchFileID", mock.Anything, &inodes.ExampleAliceFile, files.ExampleFile1.ID()).
			Return(&inodes.ExampleAliceFile, nil).Once()

		filesMock.On("Delete", mock.Anything, *inodes.ExampleAliceFile.FileID()).Return(nil).Once()

		err := runner.Run(ctx, json.RawMessage(`{
			"existing-file-id": "66278d2b-7a4f-4764-ac8a-fc08f224eb66",
			"duplicate-file-id": "abf05a02-8af9-4184-a46d-847f7d951c6b"
		}`))
		assert.NoError(t, err)
	})

	t.Run("Run with an invalid json", func(t *testing.T) {
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		runner := NewFSRemoveDuplicateFileRunner(inodesMock, filesMock, schedulerMock)

		err := runner.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})
}
