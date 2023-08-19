package oauthclients

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestOauthClientsService(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	client := Client{
		id:             "some-id",
		name:           "some-name",
		secret:         "some-secret-uuid",
		redirectURI:    "http://some-url",
		userID:         "some-user-id",
		createdAt:      now,
		scopes:         Scopes{"foo", "bar"},
		public:         true,
		skipValidation: true,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, "some-id").Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, &client).Return(nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &client, res)
	})

	t.Run("Create with a client id already taken", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, "some-id").Return(&Client{ /* some fields */ }, nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})

		assert.ErrorIs(t, err, ErrClientIDTaken)
		assert.Nil(t, res)
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, "some-id").Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, mock.Anything).Return(fmt.Errorf("some-error")).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})

		assert.EqualError(t, err, "failed to save the client: some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, "some-id").Return(&client, nil).Once()

		res, err := svc.GetByID(ctx, "some-id")
		assert.NoError(t, err)
		assert.EqualValues(t, &client, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, "some-id").Return(nil, nil).Once()

		res, err := svc.GetByID(ctx, "some-id")
		assert.NoError(t, err)
		assert.Nil(t, nil, res)
	})

	t.Run("GetByID with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, "some-id").Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetByID(ctx, "some-id")
		assert.Nil(t, nil, res)
		assert.EqualError(t, err, "failed to get by ID: some-error")
	})
}
