package dfs

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func Test_LocalFS(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("Get success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", ExampleAliceDir.ID()).Return(&ExampleAliceFile, nil).Once()

		res, err := spaceFS.Get(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"})
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceFile, res)
	})

	t.Run("Get with an element not found", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()

		res, err := spaceFS.Get(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo"})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("Get with storage error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := spaceFS.Get(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("CreateDir success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		// Call twice: The first one for the walk function in order to check if this is a directory, the second one by createDir
		// in order to check if the directory already exists.
		storageMock.On("GetByNameAndParent", mock.Anything, "new-dir", ExampleAliceRoot.ID()).Return(nil, errNotFound).Twice()

		toolsMock.ClockMock.On("Now").Return(ExampleAliceEmptyDir.lastModifiedAt.UTC()).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceEmptyDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &ExampleAliceEmptyDir).Return(nil).Once()

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/new-dir",
			CreatedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceEmptyDir, res)
	})

	t.Run("CreateDir with a validation error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/some-dir-name",
			CreatedBy: nil,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "CreatedBy: cannot be blank")
	})

	t.Run("CreateDir with an already existing file/directory", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/some-dir-name",
			CreatedBy: &users.ExampleAlice,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
	})

	t.Run("CreateDir with a GetByNameAndParent error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", ExampleAliceRoot.ID()).
			Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/some-dir-name",
			CreatedBy: &users.ExampleAlice,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("CreateDir with / as path", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := spaceFS.CreateDir(ctx, &CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "/",
			CreatedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("Remove success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := spaceFS.Remove(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"})
		assert.NoError(t, err)
	})

	t.Run("Remove with an inode not found", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errs.ErrNotFound).Once()

		err := spaceFS.Remove(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"})
		assert.NoError(t, err)
	})

	t.Run("Remove with a Get error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Remove(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Remove with a Patch error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(fmt.Errorf("some-error")).Once()

		err := spaceFS.Remove(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Remove with a RegisterFSRefreshSizeTask error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.Parent(),
			ModifiedAt: now,
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Remove(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("ListDir success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		storageMock.On("GetAllChildrens", mock.Anything, ExampleAliceDir.ID(), &storage.PaginateCmd{Limit: 2}).
			Return([]INode{ExampleAliceFile}, nil).Once()

		res, err := spaceFS.ListDir(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"}, &storage.PaginateCmd{Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, []INode{ExampleAliceFile}, res)
	})

	t.Run("ListDir with an invalid path", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errs.ErrNotFound).Once()

		res, err := spaceFS.ListDir(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"}, &storage.PaginateCmd{Limit: 2})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("ListDir with a Get error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		res, err := spaceFS.ListDir(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"}, &storage.PaginateCmd{Limit: 2})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("ListDir with a GetAllChildrens errors", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		storageMock.On("GetAllChildrens", mock.Anything, ExampleAliceDir.ID(), &storage.PaginateCmd{Limit: 2}).
			Return(nil, fmt.Errorf("some-error")).Once()

		res, err := spaceFS.ListDir(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "foo"}, &storage.PaginateCmd{Limit: 2})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Download success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		file, err := afero.TempFile(afero.NewMemMapFs(), "foo", "")
		require.NoError(t, err)

		// Get /foo/bar.txt
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar.txt", ExampleAliceDir.ID()).Return(&ExampleAliceFile, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, *ExampleAliceFile.FileID()).Return(&files.ExampleFile1, nil).Once()

		filesMock.On("Download", mock.Anything, &files.ExampleFile1).Return(file, nil).Once()

		res, err := spaceFS.Download(ctx, &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar.txt"})
		assert.NoError(t, err)
		assert.Equal(t, file, res)
	})

	t.Run("Upload success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(&files.ExampleFile1, nil).Once()
		toolsMock.ClockMock.On("Now").Return(ExampleAliceNewFile.createdAt).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceNewFile.ID()).Once()

		storageMock.On("Save", mock.Anything, &ExampleAliceNewFile).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceNewFile.ID(),
			ModifiedAt: ExampleAliceNewFile.createdAt,
		}).Return(nil).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/new.pdf",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
	})

	t.Run("Upload with a validation error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/bar.txt",
			Content:    nil,
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.EqualError(t, err, "validation: Content: cannot be blank.")
	})

	t.Run("Upload with a non existing directory", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errs.ErrNotFound).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/new.pdf",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("Upload with a file upload error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/new.pdf",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Upload with a Save error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(&files.ExampleFile1, nil).Once()
		toolsMock.ClockMock.On("Now").Return(ExampleAliceNewFile.createdAt).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceNewFile.ID()).Once()

		storageMock.On("Save", mock.Anything, &ExampleAliceNewFile).Return(fmt.Errorf("some-error")).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/new.pdf",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Upload with a RegisterFSRefreshSizeTask", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		content := "Hello, World!"

		// Get /foo
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&ExampleAliceDir, nil).Once()

		filesMock.On("Upload", mock.Anything, bytes.NewBufferString(content)).Return(&files.ExampleFile1, nil).Once()
		toolsMock.ClockMock.On("Now").Return(ExampleAliceNewFile.createdAt).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceNewFile.ID()).Once()

		storageMock.On("Save", mock.Anything, &ExampleAliceNewFile).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceNewFile.ID(),
			ModifiedAt: ExampleAliceNewFile.createdAt,
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Upload(ctx, &UploadCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			FilePath:   "foo/new.pdf",
			Content:    bytes.NewBufferString(content),
			UploadedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Move success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo.txt
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo.txt", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		schedulerMock.On("RegisterFSMoveTask", mock.Anything, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		}).Return(nil).Once()

		err := spaceFS.Move(ctx, &MoveCmd{
			Src:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo.txt"},
			Dst:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"},
			MovedBy: &users.ExampleAlice,
		})
		assert.NoError(t, err)
	})

	t.Run("Move with a validation error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		err := spaceFS.Move(ctx, &MoveCmd{
			Src:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: ""},
			Dst:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"},
			MovedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.EqualError(t, err, "validation: Src: (Path: cannot be blank.).")
	})

	t.Run("Move with a source not found", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo.txt
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo.txt", ExampleAliceRoot.ID()).Return(nil, errs.ErrNotFound).Once()

		err := spaceFS.Move(ctx, &MoveCmd{
			Src:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo.txt"},
			Dst:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"},
			MovedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("Move with a move error", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		// Get /foo.txt
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo.txt", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		schedulerMock.On("RegisterFSMoveTask", mock.Anything, &scheduler.FSMoveArgs{
			SpaceID:     spaces.ExampleAlicePersonalSpace.ID(),
			SourceInode: ExampleAliceFile.ID(),
			TargetPath:  "/bar.txt",
			MovedAt:     now,
			MovedBy:     users.ExampleAlice.ID(),
		}).Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := spaceFS.Move(ctx, &MoveCmd{
			Src:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo.txt"},
			Dst:     &PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/bar.txt"},
			MovedBy: &users.ExampleAlice,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Rename success", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetByNameAndParent", mock.Anything, "foobar.jpg", *ExampleAliceFile.Parent()).Return(nil, errNotFound).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"last_modified_at": now,
			"name":             "foobar.jpg",
		}).Return(nil).Once()

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "foobar.jpg")

		assert.NoError(t, err)
		assert.NotEqual(t, &ExampleAliceRenamedFile, res)
		assert.Equal(t, res.Name(), "foobar.jpg")
		assert.Equal(t, res.LastModifiedAt(), now)
	})

	t.Run("Rename with an empty name", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "can't be empty")
	})

	t.Run("Rename with a root inode", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		res, err := spaceFS.Rename(ctx, &ExampleAliceRoot, "foo")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "can't rename the root")
	})

	t.Run("Rename with a file with the same name", func(t *testing.T) {
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		toolsMock := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spaceFS := newLocalFS(storageMock, filesMock, &spaces.ExampleAlicePersonalSpace, spacesMock, schedulerMock, toolsMock)

		storageMock.On("GetByNameAndParent", mock.Anything, "foobar.pdf", *ExampleAliceFile.Parent()).Return(&ExampleAliceFile, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foobar (1).pdf", *ExampleAliceFile.Parent()).Return(nil, errNotFound).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"last_modified_at": now,
			"name":             "foobar (1).pdf",
		}).Return(nil).Once()

		res, err := spaceFS.Rename(ctx, &ExampleAliceFile, "foobar.pdf")
		assert.NoError(t, err)
		assert.NotEqual(t, &ExampleAliceRenamedFile, res)
		assert.Equal(t, res.Name(), "foobar (1).pdf")
		assert.Equal(t, res.LastModifiedAt(), now)
	})
}
