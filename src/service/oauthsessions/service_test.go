package oauthsessions

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/tools"
)

func Test_OauthSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

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
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceSession, res)
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

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
		assert.EqualError(t, err, "failed to save the refresh session: some-error")
	})

	t.Run("GetByAccessToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByAccessToken", mock.Anything, ExampleAliceSession.accessToken).Return(&ExampleAliceSession, nil).Once()

		res, err := service.GetByAccessToken(ctx, ExampleAliceSession.accessToken)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByRefreshToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByRefreshToken", mock.Anything, ExampleAliceSession.refreshToken).Return(&ExampleAliceSession, nil).Once()

		res, err := service.GetByRefreshToken(ctx, ExampleAliceSession.refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("RemoveByAccessToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("RemoveByAccessToken", mock.Anything, "some-access-token").Return(nil).Once()

		service.RemoveByAccessToken(ctx, "some-access-token")
	})

	t.Run("RemoveByRefreshToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("RemoveByRefreshToken", mock.Anything, "some-refresh-token").Return(nil).Once()

		service.RemoveByRefreshToken(ctx, "some-refresh-token")
	})
}
