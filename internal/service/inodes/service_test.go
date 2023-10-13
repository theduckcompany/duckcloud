package inodes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
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
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/some-dir-name",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("CreateDir success 2", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("CreateDir with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     "some-invalid-root",
			FullName: "/some-dir-name",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Root: must be a valid UUID v4.")
	})

	t.Run("CreateDir with a parent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "unknown", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/unknown/some-dir-name", // invalid path
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("CreateDir with a file as child", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: ptr.To(uuid.UUID("be73b0ad-3fef-492a-baa0-39d7964f87ab")), // File and not directory here <-,
			// some other unused fields
		}, nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
		assert.Nil(t, res)
	})

	t.Run("CreateDir with a already existing file/directory", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleBobRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/some-dir-name",
		})

		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.GetByID(ctx, ExampleAliceRoot.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		invalidID := uuid.UUID("f092f39a-1b5b-488c-8679-75607e798502")

		storageMock.On("GetByID", mock.Anything, invalidID).Return(nil, errNotFound).Once()

		res, err := service.GetByID(ctx, invalidID)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		storageMock.On("Patch", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.NoError(t, err)
	})

	t.Run("RemoveAll with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/foo",
		})

		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Root: must be a valid UUID v4.")
	})

	t.Run("RemoveAll with a file not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.NoError(t, err)
	})

	t.Run("RemoveAll with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		storageMock.On("Patch", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(errors.New("some-error")).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Get success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := INode{
			id:     uuid.UUID("eec51147-ec64-4640-b148-aceadbcb876e"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			name:   "bar",
			fileID: ptr.To(uuid.UUID("672a0dee-f6fc-42c4-9fcc-35dc911f10dd")),
			// some other unused fields
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&inode, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &inode, res)
	})

	t.Run("Get with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Get(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Root: must be a valid UUID v4.")
	})

	t.Run("Get with an invalid root", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, ErrInvalidRoot)
	})

	t.Run("GetAllDeleted success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceRoot}, nil).Once()

		res, err := service.GetAllDeleted(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, ExampleAliceRoot, res[0])
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleAliceFile.ID())
		assert.NoError(t, err)
	})

	t.Run("Readdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		paginateCmd := storage.PaginateCmd{Limit: 2}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "bar",
			fileID: nil,
			// some other unused fields
		}, nil).Once()

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

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		}, &paginateCmd)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, child1, res[0])
		assert.Equal(t, child2, res[1])
	})

	t.Run("Readdir with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/some-dir-name",
		}, &storage.PaginateCmd{Limit: 10})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Root: must be a valid UUID v4.")
	})

	t.Run("CreateFile success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := INode{
			id:             uuid.UUID("some-inode-id"),
			parent:         ptr.To(ExampleAliceRoot.ID()),
			name:           "foobar",
			size:           42,
			checksum:       "deadbeef",
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         ptr.To(uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a")),
		}

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-inode-id")).Once()

		storageMock.On("Save", mock.Anything, &inode).Return(nil).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent:     ExampleAliceRoot.ID(),
			Name:       "foobar",
			Size:       42,
			Checksum:   "deadbeef",
			UploadedAt: now,
			FileID:     uuid.UUID("b30f1f80-d07a-4c17-a543-71503624fa3a"),
		})

		assert.NoError(t, err)
		assert.Equal(t, &inode, res)
	})

	t.Run("CreateFile with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent:   "some-invalid-id",
			Name:     "foobar",
			Size:     uint64(ExampleAliceFile.Size()),
			Checksum: ExampleAliceFile.Checksum(),
			FileID:   *ExampleAliceFile.FileID(),
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Parent: must be a valid UUID v4.")
	})

	t.Run("CreateFile with a non existing parent", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent:   ExampleAliceRoot.ID(),
			Name:     "foobar",
			Size:     uint64(ExampleAliceFile.Size()),
			Checksum: ExampleAliceFile.Checksum(),
			FileID:   *ExampleAliceFile.FileID(),
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorContains(t, err, "invalid parent")
	})

	t.Run("Get with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			name:   "foo",
			fileID: ptr.To(uuid.UUID("55b12e7c-fac5-455c-a083-4e8989797d9f")), // Should be a directory with a "bar" as child
			// some other unused fields
		}, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
	})

	t.Run("CreateRootDir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			parent:         nil,
			createdAt:      now,
			lastModifiedAt: now,
			fileID:         nil,
		}

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateRootDir(ctx)

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("RegisterWrite success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now().UTC()

		storageMock.On("Patch", mock.Anything, ExampleAliceFile.id, map[string]any{
			"last_modified_at": now,
			"size":             uint64(52), // 42 + 10
		}).Return(nil).Once()

		// Duplicate in order to avoid side effects on other tests
		aliceFile := ExampleAliceFile

		err := service.RegisterWrite(ctx, &aliceFile, 10, now)
		assert.NoError(t, err)
	})

	t.Run("MkdirAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

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

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, errNotFound).Once()

		// CreateDir("/foo") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, errNotFound).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(fooDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &fooDir).Return(nil).Once()

		// Check if the folder /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()

		// CreateDir("/foo/bar") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, fooDir.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a folder already existing", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

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

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&fooDir, nil).Once()

		// Check if the folder /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()

		// CreateDir("/foo/bar") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, fooDir.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, errNotFound).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a an existing file", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceFile, nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrIsNotDir)
	})
}
