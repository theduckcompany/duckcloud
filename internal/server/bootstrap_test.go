package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func Test_bootstrap(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return([]users.User{}, nil).Once()
		userMock.On("Bootstrap", mock.Anything).Return(&users.ExampleAlice, nil).Once()
		spacesMock.On("Bootstrap", mock.Anything, &users.ExampleAlice).Return(nil).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.NoError(t, err)
	})

	t.Run("with a admin user already created", func(t *testing.T) {
		// This can append if there is an error during the space bootstrap and a restart is done.
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return([]users.User{users.ExampleAlice}, nil).Once()
		spacesMock.On("Bootstrap", mock.Anything, &users.ExampleAlice).Return(nil).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.NoError(t, err)
	})

	t.Run("with a user already created but not admin", func(t *testing.T) {
		// This case should never appear as the service forbid to remove the last admin account.
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return([]users.User{users.ExampleBob}, nil).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "No admin found")
	})

	t.Run("with a GetAll error", func(t *testing.T) {
		// This case should never appear as the service forbid to remove the last admin account.
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return(nil, errs.ErrInternal).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.ErrorIs(t, err, errs.ErrInternal)
	})
	t.Run("with a users Bootstrap error", func(t *testing.T) {
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return([]users.User{}, nil).Once()
		userMock.On("Bootstrap", mock.Anything).Return(nil, errs.ErrInternal).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("with a space bootstrap error", func(t *testing.T) {
		userMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)

		userMock.On("GetAll", mock.Anything, &storage.PaginateCmd{Limit: 4}).Return([]users.User{}, nil).Once()
		userMock.On("Bootstrap", mock.Anything).Return(&users.ExampleAlice, nil).Once()
		spacesMock.On("Bootstrap", mock.Anything, &users.ExampleAlice).Return(errs.ErrInternal).Once()

		err := bootstrap(ctx, userMock, spacesMock)
		require.ErrorIs(t, err, errs.ErrInternal)
	})
}
