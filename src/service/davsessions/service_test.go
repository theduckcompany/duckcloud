package davsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestDavSessionsService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			UserID:   users.ExampleAlice.ID(),
			FullName: "/",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-password")).Once()
		tools.UUIDMock.On("New").Return(ExampleAliceSession.ID()).Once()
		tools.ClockMock.On("Now").Return(ExampleAliceSession.createdAt).Once()

		storageMock.On("Save", mock.Anything, &ExampleAliceSession).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		res, err := service.Create(ctx, &CreateCmd{
			UserID: uuid.UUID("some-invalid-id"),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Create with an user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(nil, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: userID: not found")
	})

	t.Run("Create with a rootFS not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			UserID:   users.ExampleAlice.ID(),
			FullName: "/",
		}).Return(nil, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: rootFS: not found")
	})

	t.Run("Create with a rootFs not owner by the given user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			UserID:   users.ExampleAlice.ID(),    // Alice user id
			Root:     inodes.ExampleBobRoot.ID(), // Bob fs root
			FullName: "/",
		}).Return(nil, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleBobRoot.ID(),
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: rootFS: not found")
	})
}
