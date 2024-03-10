package davsessions

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDavSessionsService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		spacesMock.On("GetUserSpace", mock.Anything, ExampleAliceSession.userID, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-password")).Once()
		tools.UUIDMock.On("New").Return(ExampleAliceSession.ID()).Once()
		tools.ClockMock.On("Now").Return(ExampleAliceSession.createdAt).Once()

		storageMock.On("Save", mock.Anything, &ExampleAliceSession).Return(nil).Once()

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     ExampleAliceSession.Name(),
			UserID:   ExampleAliceSession.userID,
			Username: ExampleAliceSession.username,
			SpaceID:  spaces.ExampleAlicePersonalSpace.ID(),
		})

		assert.NotEmpty(t, secret)
		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     ExampleAliceSession.Name(),
			UserID:   "some-invali-id",
			Username: ExampleAliceSession.username,
			SpaceID:  spaces.ExampleAlicePersonalSpace.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Create with a space not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		spacesMock.On("GetUserSpace", mock.Anything, ExampleAliceSession.userID, spaces.ExampleAlicePersonalSpace.ID()).
			Return(nil, errNotFound).Once()

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     ExampleAliceSession.Name(),
			UserID:   ExampleAliceSession.userID,
			Username: ExampleAliceSession.username,
			SpaceID:  spaces.ExampleAlicePersonalSpace.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)

		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Create with a space not owned by the given user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		spacesMock.On("GetUserSpace", mock.Anything, ExampleAliceSession.userID, spaces.ExampleBobPersonalSpace.ID()).
			Return(&spaces.ExampleBobPersonalSpace, nil).Once()

		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     ExampleAliceSession.Name(),
			UserID:   ExampleAliceSession.userID,
			Username: ExampleAliceSession.username,
			SpaceID:  spaces.ExampleBobPersonalSpace.ID(),
		})

		assert.Nil(t, res)
		assert.Empty(t, secret)
		require.EqualError(t, err, "bad request: invalid spaceID")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.id, &storage.PaginateCmd{Limit: 10}).Return([]DavSession{ExampleAliceSession}, nil).Once()

		res, err := service.GetAllForUser(ctx, ExampleAliceSession.id, &storage.PaginateCmd{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, []DavSession{ExampleAliceSession}, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetByUsernameAndPassword", mock.Anything, "some-username", secret.NewText(hex.EncodeToString([]byte("some-password")))).
			Return(&ExampleAliceSession, nil).Once()

		res, err := service.Authenticate(ctx, "some-username", secret.NewText("some-password"))
		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: ExampleAliceSession.ID(),
		})
		require.NoError(t, err)
	})

	t.Run("Delete with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: "some invalid id",
		})
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "SessionID: must be a valid UUID v4.")
	})

	t.Run("Delete with a session not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(nil, errNotFound).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    ExampleAliceSession.UserID(),
			SessionID: ExampleAliceSession.ID(),
		})
		require.NoError(t, err)
	})

	t.Run("Delete with a session owner by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID:    uuid.UUID("de946548-095f-4fa9-8f03-81d3459f8000"), // some random id
			SessionID: ExampleAliceSession.ID(),
		})
		require.EqualError(t, err, "not found: user ids are not matching")
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return([]DavSession{ExampleAliceSession}, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(nil).Once()

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		require.NoError(t, err)
	})

	t.Run("DeleteAll with a GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.UserID(), (*storage.PaginateCmd)(nil)).Return([]DavSession{ExampleAliceSession, ExampleAliceSession}, nil).Once()
		storageMock.On("GetByID", mock.Anything, ExampleAliceSession.ID()).Return(&ExampleAliceSession, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, ExampleAliceSession.ID()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByID and RemoveByID a second time

		err := service.DeleteAll(ctx, ExampleAliceSession.UserID())
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
