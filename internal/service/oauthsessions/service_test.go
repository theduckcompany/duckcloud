package oauthsessions

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func Test_OauthSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		tools.ClockMock.On("Now").Return(ExampleAliceSession.accessCreatedAt).Once()
		storageMock.On("Save", mock.Anything, &ExampleAliceSession).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			AccessToken:      ExampleAliceSession.AccessToken(),
			AccessExpiresAt:  ExampleAliceSession.AccessExpiresAt(),
			RefreshToken:     ExampleAliceSession.RefreshToken(),
			RefreshExpiresAt: ExampleAliceSession.RefreshExpiresAt(),
			ClientID:         ExampleAliceSession.ClientID(),
			UserID:           ExampleAliceSession.UserID(),
			Scope:            ExampleAliceSession.Scope(),
		})
		require.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceSession, res)
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		tools.ClockMock.On("Now").Return(ExampleAliceSession.accessCreatedAt).Once()
		storageMock.On("Save", mock.Anything, &ExampleAliceSession).Return(fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			AccessToken:      ExampleAliceSession.AccessToken(),
			AccessExpiresAt:  ExampleAliceSession.AccessExpiresAt(),
			RefreshToken:     ExampleAliceSession.RefreshToken(),
			RefreshExpiresAt: ExampleAliceSession.RefreshExpiresAt(),
			ClientID:         ExampleAliceSession.ClientID(),
			UserID:           ExampleAliceSession.UserID(),
			Scope:            ExampleAliceSession.Scope(),
		})
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("GetByAccessToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("GetByAccessToken", mock.Anything, ExampleAliceSession.accessToken).Return(&ExampleAliceSession, nil).Once()

		res, err := service.GetByAccessToken(ctx, ExampleAliceSession.accessToken)
		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByRefreshToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("GetByRefreshToken", mock.Anything, ExampleAliceSession.refreshToken).Return(&ExampleAliceSession, nil).Once()

		res, err := service.GetByRefreshToken(ctx, ExampleAliceSession.refreshToken)
		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("RemoveByAccessToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("RemoveByAccessToken", mock.Anything, secret.NewText("some-access-token")).Return(nil).Once()

		err := service.RemoveByAccessToken(ctx, secret.NewText("some-access-token"))
		require.NoError(t, err)
	})

	t.Run("RemoveByRefreshToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("RemoveByRefreshToken", mock.Anything, secret.NewText("some-refresh-token")).Return(nil).Once()

		err := service.RemoveByRefreshToken(ctx, secret.NewText("some-refresh-token"))
		require.NoError(t, err)
	})

	t.Run("DeleteAllForUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.userID, (*sqlstorage.PaginateCmd)(nil)).Return([]Session{ExampleAliceSession}, nil).Once()
		storageMock.On("RemoveByAccessToken", mock.Anything, ExampleAliceSession.accessToken).Return(nil).Once()

		err := service.DeleteAllForUser(ctx, ExampleAliceSession.userID)
		require.NoError(t, err)
	})

	t.Run("DeleteAllForUser stop directly in case of error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(tools, storageMock)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceSession.userID, (*sqlstorage.PaginateCmd)(nil)).Return([]Session{ExampleAliceSession, ExampleAliceSession}, nil).Once()
		storageMock.On("RemoveByAccessToken", mock.Anything, ExampleAliceSession.accessToken).Return(fmt.Errorf("some-error")).Once()
		// Do not call "RemoveByAccessToken" a second time for the second error

		err := service.DeleteAllForUser(ctx, ExampleAliceSession.userID)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
