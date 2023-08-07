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
	now := time.Now()

	t.Run("CreateDirectory success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		inode := &INode{
			ID:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			UserID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent:         uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			Type:           Directory,
			CreatedAt:      now,
			LastModifiedAt: now,
		}

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&INode{
			ID:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			UserID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			// some other unused fields
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storage.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateDirectory(ctx, &CreateDirectoryCmd{
			Name:   "some-dir-name",
			UserID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent: uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("CreateDirectory with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		res, err := service.CreateDirectory(ctx, &CreateDirectoryCmd{
			Name:   "some-dir-name",
			UserID: uuid.UUID("some-invalid-id"),
			Parent: uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("CreateDirectory with a parent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.CreateDirectory(ctx, &CreateDirectoryCmd{
			Name:   "some-dir-name",
			UserID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent: uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: invalid parent: parent doesn't exists")
	})

	t.Run("CreateDirectory with a parent owned by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&INode{
			ID:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			UserID: uuid.UUID("some-other-user-id"),
			// some other unused fields
		}, nil).Once()

		res, err := service.CreateDirectory(ctx, &CreateDirectoryCmd{
			Name:   "some-dir-name",
			UserID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent: uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: invalid parent: parent not authorized")
	})

	t.Run("GetByUserAndID", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		inode := &INode{
			ID:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			UserID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent:         uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			Type:           Directory,
			CreatedAt:      now,
			LastModifiedAt: now,
		}

		storage.On("GetByID", mock.Anything, uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Return(inode, nil).Once()

		res, err := service.GetByUserAndID(ctx, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"))

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("BootstrapUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		inode := &INode{
			ID:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			UserID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			Parent:         NoParent,
			Type:           Directory,
			CreatedAt:      now,
			LastModifiedAt: now,
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
