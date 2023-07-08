package oauthclients

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOauthClientsService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		now := time.Now()

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, "some-id").Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, &Client{
			ID:             "some-id",
			Name:           "some-name",
			Secret:         "some-secret-uuid",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			CreatedAt:      now,
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		}).Return(nil).Once()

		err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})
		assert.NoError(t, err)
	})

	t.Run("Create with a client id already taken", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByID", mock.Anything, "some-id").Return(&Client{ /* some fields */ }, nil).Once()

		err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})
		assert.ErrorIs(t, err, ErrClientIDTaken)
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		now := time.Now()

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, "some-id").Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, mock.Anything).Return(fmt.Errorf("some-error")).Once()

		err := svc.Create(ctx, &CreateCmd{
			ID:             "some-id",
			Name:           "some-name",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		})
		assert.EqualError(t, err, "failed to save the client: some-error")
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		client := Client{
			ID:             "some-id",
			Name:           "some-name",
			Secret:         "some-secret-uuid",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			CreatedAt:      time.Now(),
			Scopes:         Scopes{"foo", "bar"},
			Public:         true,
			SkipValidation: true,
		}

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
