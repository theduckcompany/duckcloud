package inodes

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestInodes(t *testing.T) {
	ctx := context.Background()

	t.Run("Mkdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

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

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storage.On("Save", mock.Anything, inode).Return(nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

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

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "foo", ExampleRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent: ExampleRoot.ID(),
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storage.On("Save", mock.Anything, inode).Return(nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "unknown", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), "foo", ExampleRoot.ID()).Return(&INode{
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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		userID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

		inode := INode{
			id:       uuid.UUID("eec51147-ec64-4640-b148-aceadbcb876e"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			nodeType: File,
			name:     "bar",
			// some other unused fields
		}

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, userID, "foo", ExampleRoot.ID()).Return(&INode{
			id:       uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			userID:   uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:   ExampleRoot.ID(),
			nodeType: Directory,
			name:     "foo",
			// some other unused fields
		}, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, userID, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&inode, nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		res, err := service.Open(ctx, &PathCmd{
			Root:     ExampleRoot.ID(),
			FullName: "/foo/bar",
			UserID:   uuid.UUID("d35f9848-6310-4280-bc9a-44534035a401"), // UserID != ExampleRoot.UserID
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "not found: dir \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" is not owned by \"d35f9848-6310-4280-bc9a-44534035a401\"")
	})

	t.Run("Open with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		userID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleRoot, nil).Once()

		storage.On("GetByNameAndParent", mock.Anything, userID, "foo", ExampleRoot.ID()).Return(&INode{
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
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			parent:         NoParent,
			nodeType:       Directory,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storage.On("CountUserINodes", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")).Return(uint(0), nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storage.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.BootstrapUser(ctx, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"))

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("BootstrapUser with an already bootstraped fs", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("CountUserINodes", mock.Anything, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")).Return(uint(2), nil).Once()

		res, err := service.BootstrapUser(ctx, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"))

		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: this user is already bootstraped")
	})
}
