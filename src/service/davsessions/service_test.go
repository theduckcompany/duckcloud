package davsessions

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
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

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:   ExampleAliceSession.Name(),
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.NotEmpty(t, secret)
		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:   ExampleAliceSession.Name(),
			UserID: uuid.UUID("some-invalid-id"),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Create with an user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(storageMock, inodesMock, usersMock, tools)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(nil, nil).Once()

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:   ExampleAliceSession.Name(),
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
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

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:   ExampleAliceSession.Name(),
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleAliceRoot.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
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

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:   ExampleAliceSession.Name(),
			UserID: users.ExampleAlice.ID(),
			FSRoot: inodes.ExampleBobRoot.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
		assert.EqualError(t, err, "validation error: rootFS: not found")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.id, &storage.PaginateCmd{Limit: 10}).Return([]DavSession{ExampleAliceSession}, nil).Once()

		res, err := service.GetAllForUser(ctx, ExampleAliceSession.id, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []DavSession{ExampleAliceSession}, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		hashedPasswd := "f0ce9d6e7315534d2f3603d11f496dafcda25f2f5bc2b4f8292a8ee34fe7735b" // sha256 of "some-password"

		storageMock.On("GetByUsernameAndPassHash", mock.Anything, "some-username", hashedPasswd).Return(&ExampleAliceSession, nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: ExampleAliceSession.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("Delete with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: "some invalid id",
		})
		assert.EqualError(t, err, "validation error: SessionID: must be a valid UUID v4.")
	})

	t.Run("Delete with a session not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(nil, nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: ExampleAliceSession.ID(),
		})
		assert.NoError(t, err)
	})

	t.Run("Delete with a session owner by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    uuid.UUID("de946548-095f-4fa9-8f03-81d3459f8000"), // some random id
			SessionID: ExampleAliceSession.ID(),
		})
		assert.EqualError(t, err, "not found: user ids are not matching")
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return([]DavSession{ExampleAliceSession}, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(nil).Once()

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		assert.NoError(t, err)
	})

	t.Run("DeleteAll witha GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		assert.EqualError(t, err, "failed to GetAllForUser: some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)

		service := NewService(storageMock, inodesMock, usersMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return([]DavSession{ExampleAliceSession, ExampleAliceSession}, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByID and RemoveByID a second time

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		assert.EqualError(t, err, fmt.Sprintf("failed to Delete dav session %q: failed to RemoveByID: some-error", ExampleAliceSession.id))
	})
}
