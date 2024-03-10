package oauthclients

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestOauthClientsService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, &ExampleAliceClient).Return(nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             ExampleAliceClient.id,
			Name:           ExampleAliceClient.name,
			RedirectURI:    ExampleAliceClient.redirectURI,
			UserID:         ExampleAliceClient.userID,
			Scopes:         ExampleAliceClient.scopes,
			Public:         ExampleAliceClient.public,
			SkipValidation: ExampleAliceClient.skipValidation,
		})

		require.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceClient, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             ExampleAliceClient.id,
			Name:           ExampleAliceClient.name,
			RedirectURI:    "some-invalid-url",
			UserID:         ExampleAliceClient.userID,
			Scopes:         ExampleAliceClient.scopes,
			Public:         ExampleAliceClient.public,
			SkipValidation: ExampleAliceClient.skipValidation,
		})

		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "RedirectURI: must be a valid URL.")
		assert.Nil(t, res)
	})

	t.Run("Create with a client id already taken", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(&ExampleAliceClient, nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             ExampleAliceClient.id,
			Name:           ExampleAliceClient.name,
			RedirectURI:    ExampleAliceClient.redirectURI,
			UserID:         ExampleAliceClient.userID,
			Scopes:         ExampleAliceClient.scopes,
			Public:         ExampleAliceClient.public,
			SkipValidation: ExampleAliceClient.skipValidation,
		})

		require.ErrorIs(t, err, ErrClientIDTaken)
		assert.Nil(t, res)
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		// Check that the client name is not already taken
		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()                          // Client.CreatedAt
		tools.UUIDMock.On("New").Return(uuid.UUID("some-secret-uuid")).Once() // Client.Secret

		storage.On("Save", mock.Anything, &ExampleAliceClient).Return(fmt.Errorf("some-error")).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			ID:             ExampleAliceClient.id,
			Name:           ExampleAliceClient.name,
			RedirectURI:    ExampleAliceClient.redirectURI,
			UserID:         ExampleAliceClient.userID,
			Scopes:         ExampleAliceClient.scopes,
			Public:         ExampleAliceClient.public,
			SkipValidation: ExampleAliceClient.skipValidation,
		})

		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(&ExampleAliceClient, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAliceClient.id)
		require.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceClient, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(nil, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAliceClient.id)
		require.NoError(t, err)
		assert.Nil(t, nil, res)
	})

	t.Run("GetByID with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		svc := newService(tools, storage)

		storage.On("GetByID", mock.Anything, ExampleAliceClient.id).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetByID(ctx, ExampleAliceClient.id)
		assert.Nil(t, nil, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
