package inodes

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestInodes(t *testing.T) {
	ctx := context.Background()

	t.Run("Mkdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:         uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType:       Directory,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/some-dir-name",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("Mkdir success 2", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "bar",
			userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:         uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType:       Directory,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "foo", ExampleRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent: ExampleRoot.ID(),
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("Mkdir with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/some-dir-name",
			UserID:   uuid.UUID("some-invalid-id"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Mkdir with a parent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "unknown", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/unknown/some-dir-name", // invalid path
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "mkdir /unknown/some-dir-name: file does not exist")
	})

	t.Run("Mkdir with a parent owned by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/some-dir-name",
			UserID:   uuid.UUID("d35f9848-6310-4280-bc9a-44534035a401"), // UserID != inodes.ExampleRoot.UserID
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "not found: dir \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" is not owned by \"d35f9848-6310-4280-bc9a-44534035a401\"")
	})

	t.Run("Mkdir with a file as child", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "foo", ExampleRoot.ID()).Return(&INode{
			id:       uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   ExampleRoot.ID(),
			nodeType: File, // File and not directory here <-,
			name:     "foo",
			// some other unused fields
		}, nil).Once()

		res, err := service.Mkdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.EqualError(t, err, "mkdir /foo/bar: invalid argument")
		assert.Nil(t, res)
	})

	t.Run("Open success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		userID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

		inode := INode{
			id:       uuid.UUID("eec51147-ec64-4640-b148-aceadbcb876e"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType: File,
			name:     "bar",
			// some other unused fields
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, userID, "foo", ExampleRoot.ID()).Return(&INode{
			id:       uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   ExampleRoot.ID(),
			nodeType: Directory,
			name:     "foo",
			// some other unused fields
		}, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, userID, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&inode, nil).Once()

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &inode, res)
	})

	t.Run("Open with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("not an id"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Open with an invalid root", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "not found: root \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" not found")
	})

	t.Run("Open with a root owned by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("d35f9848-6310-4280-bc9a-44534035a401"), // UserID != ExampleRoot.UserID
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "not found: dir \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" is not owned by \"d35f9848-6310-4280-bc9a-44534035a401\"")
	})

	t.Run("Readdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		paginateCmd := storage.PaginateCmd{Limit: 2}
		userID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, userID, "foo", ExampleRoot.ID()).Return(&INode{
			id:       uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   ExampleRoot.ID(),
			nodeType: Directory,
			name:     "bar",
			// some other unused fields
		}, nil).Once()

		child1 := INode{
			id:       uuid.UUID("b3411c4b-acc3-4f79-a54e-f315a18ce6c7"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType: Directory,
			name:     "some-dir",
		}
		child2 := INode{
			id:       uuid.UUID("0af1f541-454e-4c7d-a871-706d9c5ad2cc"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType: File,
			name:     "some-file",
		}

		storageMock.On("GetAllChildrens", mock.Anything, userID, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), &paginateCmd).Return(
			[]INode{child1, child2}, nil).Once()

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
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
			Root:     ExampleRoot.ID(),
			FullName: "/some-dir-name",
			UserID:   uuid.UUID("some-invalid-id"),
		}, &storage.PaginateCmd{Limit: 10})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Readdir with a parent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "unknown", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/unknown/some-dir-name", // invalid path
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		}, &storage.PaginateCmd{Limit: 10})

		assert.Nil(t, res)
		assert.EqualError(t, err, "readdir /unknown/some-dir-name: file does not exist")
	})

	t.Run("Open with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		userID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, userID, "foo", ExampleRoot.ID()).Return(&INode{
			id:       uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   ExampleRoot.ID(),
			nodeType: File, // Should be a directory with a "bar" as child
			name:     "foo",
			// some other unused fields
		}, nil).Once()

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
	})

	t.Run("BootstrapUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:         NoParent,
			nodeType:       Directory,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("CountUserINodes", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")).Return(uint(0), nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.BootstrapUser(ctx, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"))

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("BootstrapUser with an already bootstraped fs", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("CountUserINodes", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")).Return(uint(2), nil).Once()

		res, err := service.BootstrapUser(ctx, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"))

		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: this user is already bootstraped")
	})
}
