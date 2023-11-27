package inodes

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestINodes(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateDir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now()
		inode := INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			parent:         ptr.To(ExampleAliceRoot.ID()),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storageMock.On("Save", mock.Anything, &inode).Return(nil).Once()

		res, err := service.CreateDir(ctx, &ExampleAliceRoot, "some-dir-name")
		assert.NoError(t, err)
		assert.EqualValues(t, &inode, res)
	})

	t.Run("CreateDir with an already existing file/directory", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", ExampleAliceRoot.ID()).Return(&ExampleAliceFile, nil).Once()

		res, err := service.CreateDir(ctx, &ExampleAliceRoot, "some-dir-name")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.GetByID(ctx, ExampleAliceRoot.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		invalidID := uuid.UUID("f092f39a-1b5b-488c-8679-75607e798502")

		storageMock.On("GetByID", mock.Anything, invalidID).Return(nil, errNotFound).Once()

		res, err := service.GetByID(ctx, invalidID)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Remove success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.Parent(),
			ModifiedAt: now,
		}).Return(nil).Once()

		err := service.Remove(ctx, &ExampleAliceFile)

		assert.NoError(t, err)
	})

	t.Run("Remove with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(errors.New("some-error")).Once()

		err := service.Remove(ctx, &ExampleAliceFile)

		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Remove with a task error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      *ExampleAliceDir.Parent(),
			ModifiedAt: now,
		}).Return(errors.New("some-error")).Once()

		err := service.Remove(ctx, &ExampleAliceFile)
		assert.EqualError(t, err, "failed to schedule the fs-refresh-size task: some-error")
	})

	t.Run("Get success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		inode := INode{
			id:     uuid.UUID("eec51147-ec64-4640-b148-aceadbcb876e"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			name:   "bar",
			fileID: ptr.To(uuid.UUID("672a0dee-f6fc-42c4-9fcc-35dc911f10dd")),
			// some other unused fields
		}

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&inode, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &inode, res)
	})

	t.Run("Get with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		res, err := service.Get(ctx, &PathCmd{
			Space: nil,
			Path:  "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Space: cannot be blank.")
	})

	t.Run("Get with an invalid root", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := service.Get(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, ErrInvalidRoot)
	})

	t.Run("Get with a child not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		res, err := service.Get(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorContains(t, err, `"/foo" doesn't have a child named "bar"`)
	})

	t.Run("GetAllDeleted success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceRoot}, nil).Once()

		res, err := service.GetAllDeleted(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, ExampleAliceRoot, res[0])
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, &ExampleAliceFile)
		assert.NoError(t, err)
	})

	t.Run("Readdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		paginateCmd := storage.PaginateCmd{Limit: 2}

		child1 := INode{
			id:     uuid.UUID("b3411c4b-acc3-4f79-a54e-f315a18ce6c7"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			name:   "some-dir",
			fileID: nil,
		}
		child2 := INode{
			id:     uuid.UUID("0af1f541-454e-4c7d-a871-706d9c5ad2cc"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			name:   "some-file",
			fileID: ptr.To(uuid.UUID("7680ca50-2312-42d2-99d9-022d8879b7ec")),
		}

		storageMock.On("GetAllChildrens", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), &paginateCmd).Return(
			[]INode{child1, child2}, nil).Once()

		res, err := service.Readdir(ctx, &ExampleAliceRoot, &paginateCmd)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, child1, res[0])
		assert.Equal(t, child2, res[1])
	})

	t.Run("CreateFile success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		inode := INode{
			id:             uuid.UUID("some-inode-id"),
			parent:         ptr.To(ExampleAliceRoot.ID()),
			name:           "foobar",
			spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
			size:           0,
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         ptr.To(uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a")),
		}

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-inode-id")).Once()

		storageMock.On("Save", mock.Anything, &inode).Return(nil).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			Parent:     ExampleAliceRoot.ID(),
			Name:       "foobar",
			UploadedAt: now,
			FileID:     uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a"),
		})

		assert.NoError(t, err)
		assert.Equal(t, &inode, res)
	})

	t.Run("CreateFile with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			Parent:     "some-invalid-id",
			Name:       "foobar",
			FileID:     *ExampleAliceFile.FileID(),
			UploadedAt: now,
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Parent: must be a valid UUID v4.")
	})

	t.Run("CreateFile with a non existing parent", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Space:      &spaces.ExampleAlicePersonalSpace,
			Parent:     ExampleAliceRoot.ID(),
			Name:       "foobar",
			FileID:     *ExampleAliceFile.FileID(),
			UploadedAt: now,
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorContains(t, err, "invalid parent")
	})

	t.Run("Get with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: ptr.To(uuid.UUID("55b12e7c-fac5-455c-a083-4e8989797d9f")), // Should be a directory with a "bar" as child
			// some other unused fields
		}, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
	})

	t.Run("CreateRootDir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			parent:         nil,
			spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateRootDir(ctx, &spaces.ExampleAlicePersonalSpace)

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("MkdirAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now().UTC()
		fooDir := INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "foo",
			parent:         ptr.To(ExampleAliceRoot.ID()),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}
		barDir := INode{
			id:             uuid.UUID("1afc4ef3-d0e8-4efe-8e37-4d23acc5df9c"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		// Check if the space /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		// CreateDir("/foo") internal
		// Check if the space already exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()
		// Generate an save ""/foo/bar"" space
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(fooDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &fooDir).Return(nil).Once()

		// Check if the space /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()

		// CreateDir("/foo/bar") internal
		// Check if the space already exists
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()
		// Generate an save ""/foo/bar"" space
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a space already existing", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		now := time.Now().UTC()
		fooDir := INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "foo",
			parent:         ptr.To(ExampleAliceRoot.ID()),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}
		barDir := INode{
			id:             uuid.UUID("1afc4ef3-d0e8-4efe-8e37-4d23acc5df9c"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		// Check if the space /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&fooDir, nil).Once()

		// Check if the space /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()

		// CreateDir("/foo/bar") internal
		// Check if the space already exists
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()

		// Generate an save ""/foo/bar"" space
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a an existing file", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		// Check if the space /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceFile, nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
	})

	t.Run("MkdirAll with / as fullname", func(t *testing.T) {
		// In this case it will return the Root.
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Space: &spaces.ExampleAlicePersonalSpace,
			Path:  "/",
		})

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceRoot, res)
	})

	t.Run("PatchMove success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("Patch", mock.Anything, ExampleAliceFile.id, map[string]interface{}{
			"name":             "file.txt",
			"parent":           ExampleAliceDir.id,
			"last_modified_at": now,
		}).Return(nil).Once()

		res, err := service.PatchMove(ctx, &ExampleAliceFile, &ExampleAliceDir, "file.txt", now)
		assert.NoError(t, err)
		assert.Equal(t, "file.txt", res.name)
		assert.Equal(t, now, res.lastModifiedAt)
		assert.Equal(t, &ExampleAliceDir.id, res.parent)
	})

	t.Run("Move with a Patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		// Update the source parent an name
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.id, map[string]interface{}{
			"name":             "file.txt",
			"parent":           ExampleAliceDir.id,
			"last_modified_at": now,
		}).Return(fmt.Errorf("some-error")).Once()

		res, err := service.PatchMove(ctx, &ExampleAliceFile, &ExampleAliceDir, "file.txt", now)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetSumChildsSize success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSumChildsSize", mock.Anything, uuid.UUID("some-id")).Return(uint64(1024), nil).Once()

		res, err := service.GetSumChildsSize(ctx, uuid.UUID("some-id"))
		assert.NoError(t, err)
		assert.Equal(t, uint64(1024), res)
	})

	t.Run("RegisterModification success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"last_modified_at": now,
			"size":             uint64(24),
		}).Return(nil).Once()

		err := service.RegisterModification(ctx, &ExampleAliceFile, 24, now)
		assert.NoError(t, err)
	})

	t.Run("PatchFileID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"file_id": uuid.UUID("some-new-id"),
		}).Return(nil).Once()

		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceFile.ID(),
			ModifiedAt: ExampleAliceDir.LastModifiedAt(),
		}).Return(nil).Once()

		input := ExampleAliceFile

		res, err := service.PatchFileID(ctx, &input, uuid.UUID("some-new-id"))
		assert.NoError(t, err)

		expected := ExampleAliceFile
		expected.fileID = ptr.To(uuid.UUID("some-new-id"))

		assert.Equal(t, &expected, res)
	})

	t.Run("PatchFileID with a patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("Patch", mock.Anything, ExampleAliceFile.ID(), map[string]any{
			"file_id": uuid.UUID("some-new-id"),
		}).Return(errors.New("some-error")).Once()

		input := ExampleAliceFile

		res, err := service.PatchFileID(ctx, &input, uuid.UUID("some-new-id"))
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetAllInodesWithFile success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetAllInodesWithFileID", mock.Anything, uuid.UUID("some-file-id")).
			Return([]INode{ExampleAliceFile}, nil).Once()

		res, err := service.GetAllInodesWithFileID(ctx, uuid.UUID("some-file-id"))
		assert.NoError(t, err)
		assert.Equal(t, []INode{ExampleAliceFile}, res)
	})

	t.Run("GetSpaceRoot success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.GetSpaceRoot(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceRoot, res)
	})

	t.Run("GetSpaceRoot not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := service.GetSpaceRoot(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.Nil(t, res)
	})

	t.Run("GetSpaceRoot with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(schedulerMock, tools, storageMock)

		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.GetSpaceRoot(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})
}
