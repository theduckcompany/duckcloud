package davsessions

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDavSessionsService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		now := time.Now().UTC()
		user := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).WithOwners(*user).Build()
		sessionPassword := "some-password"
		session := NewFakeSession(t).
			WithName("My Session").
			WithUsername("some-username").
			WithSpace(space).
			WithPassword(sessionPassword).
			CreatedAt(now).
			CreatedBy(user).
			Build()

		// Mocks
		spacesMock.On("GetUserSpace", mock.Anything, user.ID(), space.ID()).Return(space, nil).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID(sessionPassword)).Once()
		tools.UUIDMock.On("New").Return(session.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, session).Return(nil).Once()

		// Run
		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     "My Session",
			UserID:   user.ID(),
			Username: "some-username",
			SpaceID:  space.ID(),
		})

		// Asserts
		assert.NotEmpty(t, secret)
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		space := spaces.NewFakeSpace(t).Build()
		session := NewFakeSession(t).Build()

		// Mocks

		// Run
		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     session.name,
			UserID:   "some-invali-id",
			Username: session.username,
			SpaceID:  space.ID(),
		})

		// Assets
		assert.Nil(t, res)
		assert.Empty(t, secret)
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Create with a space not found", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).WithOwners(*user).Build()
		session := NewFakeSession(t).Build()

		// Mocks
		spacesMock.On("GetUserSpace", mock.Anything, user.ID(), space.ID()).Return(nil, errNotFound).Once()

		// Run
		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     session.name,
			UserID:   user.ID(),
			Username: session.username,
			SpaceID:  space.ID(),
		})

		// Assets
		assert.Nil(t, res)
		assert.Empty(t, secret)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Create with a space not owned by the given user", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).Build() // The space doesn't have the user as owner
		session := NewFakeSession(t).Build()

		// Mocks
		spacesMock.On("GetUserSpace", mock.Anything, user.ID(), space.ID()).Return(space, nil).Once()

		// Run
		res, secret, err := service.Create(ctx, &CreateCmd{
			Name:     session.name,
			UserID:   user.ID(),
			Username: session.username,
			SpaceID:  space.ID(),
		})

		// Assets
		assert.Nil(t, res)
		assert.Empty(t, secret)
		require.EqualError(t, err, "bad request: invalid spaceID")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		session := NewFakeSession(t).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, session.id, &sqlstorage.PaginateCmd{Limit: 10}).Return([]DavSession{*session}, nil).Once()

		// Run
		res, err := service.GetAllForUser(ctx, session.id, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []DavSession{*session}, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		sessionPassword := "some-password"
		session := NewFakeSession(t).
			WithPassword(sessionPassword).
			Build()

		// Mocks
		storageMock.On("GetByUsernameAndPassword", mock.Anything, session.username, secret.NewText(hex.EncodeToString([]byte("some-password")))).
			Return(session, nil).Once()

		// Run
		res, err := service.Authenticate(ctx, session.username, secret.NewText(sessionPassword))

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		session := NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, session.ID()).Return(session, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, session.ID()).Return(nil).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID:    user.ID(),
			SessionID: session.ID(),
		})

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Delete with a validation error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID:    user.ID(),
			SessionID: "some invalid id",
		})

		// Asserts
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "SessionID: must be a valid UUID v4.")
	})

	t.Run("Delete with a session not found", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		session := NewFakeSession(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, session.ID()).Return(nil, errNotFound).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID:    user.ID(),
			SessionID: session.ID(),
		})

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Delete with a session owner by someone else", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		session := NewFakeSession(t).Build() // Session is not created by user

		// Mocks
		storageMock.On("GetByID", mock.Anything, session.ID()).Return(session, nil).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID:    user.ID(),
			SessionID: session.ID(),
		})

		// Asserts
		require.EqualError(t, err, "not found: user ids are not matching")
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		session := NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]DavSession{*session}, nil).Once()
		storageMock.On("GetByID", mock.Anything, session.ID()).Return(session, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, session.ID()).Return(nil).Once()

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("DeleteAll with a GetAll error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		spacesMock := spaces.NewMockService(t)
		service := newService(storageMock, spacesMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		session := NewFakeSession(t).CreatedBy(user).Build()
		session2 := NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]DavSession{*session, *session2}, nil).Once()
		storageMock.On("GetByID", mock.Anything, session.ID()).Return(session, nil).Once()
		storageMock.On("RemoveByID", mock.Anything, session.ID()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByID and RemoveByID a session2

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
