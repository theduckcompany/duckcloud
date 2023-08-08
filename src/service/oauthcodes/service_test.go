package oauthcodes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOauthCodeService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		now := time.Now()
		expiresAt := now.Add(time.Hour)

		tools.ClockMock.On("Now").Return(now).Once()

		storage.On("Save", mock.Anything, &Code{
			code:            "some-code",
			createdAt:       now,
			expiresAt:       expiresAt,
			clientID:        "dcca1ba7-6fa1-4684-8602-85adcb6a03a2",
			userID:          "767c0845-db3d-49df-9b14-bd4dab4dacd8",
			redirectURI:     "http://some-redirect",
			scope:           "foo,bar",
			challenge:       "some-secret",
			challengeMethod: "S256",
		}).Return(nil).Once()

		err := svc.Create(ctx, &CreateCmd{
			Code:            "some-code",
			ExpiresAt:       expiresAt,
			ClientID:        "dcca1ba7-6fa1-4684-8602-85adcb6a03a2",
			UserID:          "767c0845-db3d-49df-9b14-bd4dab4dacd8",
			RedirectURI:     "http://some-redirect",
			Scope:           "foo,bar",
			Challenge:       "some-secret",
			ChallengeMethod: "S256",
		})
		assert.NoError(t, err)
	})

	t.Run("Create with a storae error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		now := time.Now()
		expiresAt := now.Add(time.Hour)

		tools.ClockMock.On("Now").Return(now).Once()

		storage.On("Save", mock.Anything, mock.Anything).Return(fmt.Errorf("some-error")).Once()

		err := svc.Create(ctx, &CreateCmd{
			Code:            "some-code",
			ExpiresAt:       expiresAt,
			ClientID:        "dcca1ba7-6fa1-4684-8602-85adcb6a03a2",
			UserID:          "767c0845-db3d-49df-9b14-bd4dab4dacd8",
			RedirectURI:     "http://some-redirect",
			Scope:           "foo,bar",
			Challenge:       "some-secret",
			ChallengeMethod: "S256",
		})
		assert.EqualError(t, err, "failed to save the code: some-error")
	})

	t.Run("RemoveByCode", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("RemoveByCode", mock.Anything, "some-code").Return(nil).Once()

		err := svc.RemoveByCode(ctx, "some-code")
		assert.NoError(t, err)
	})

	t.Run("RemoveByCode with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("RemoveByCode", mock.Anything, "some-code").Return(fmt.Errorf("some-error")).Once()

		err := svc.RemoveByCode(ctx, "some-code")
		assert.EqualError(t, err, "some-error")
	})

	t.Run("GetByCode", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		code := Code{
			code:            "some-code",
			expiresAt:       time.Now(),
			clientID:        "dcca1ba7-6fa1-4684-8602-85adcb6a03a2",
			userID:          "767c0845-db3d-49df-9b14-bd4dab4dacd8",
			redirectURI:     "http://some-redirect",
			scope:           "foo,bar",
			challenge:       "some-secret",
			challengeMethod: "S256",
		}

		storage.On("GetByCode", mock.Anything, "some-code").Return(&code, nil).Once()

		res, err := svc.GetByCode(ctx, "some-code")
		assert.EqualValues(t, &code, res)
		assert.NoError(t, err)
	})

	t.Run("GetByCode with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		svc := NewService(tools, storage)

		storage.On("GetByCode", mock.Anything, "some-code").Return(nil, fmt.Errorf("some-error")).Once()

		code, err := svc.GetByCode(ctx, "some-code")
		assert.Nil(t, code)
		assert.EqualError(t, err, "some-error")
	})
}
