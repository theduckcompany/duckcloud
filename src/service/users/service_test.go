package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/password"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Service_Create_success(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	user := User{
		ID:        uuid.UUID("some-id"),
		Username:  "some-username",
		Email:     "some@email.com",
		password:  "some-encrypted-password",
		CreatedAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByEmail", ctx, "some@email.com").Return(nil, nil).Once()
		storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-id")).Once()
		tools.ClockMock.On("Now").Return(now)
		tools.PasswordMock.On("Encrypt", ctx, "some-password").Return("some-encrypted-password", nil).Once()

		storage.On("Save", ctx, &user).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Email:    "some@email.com",
			Password: "some-password",
		})
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("Create with email already exists", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByEmail", ctx, "some@email.com").Return(&User{}, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Email:    "some@email.com",
			Password: "some-password",
		})
		assert.ErrorIs(t, err, ErrAlreadyExists)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.Nil(t, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByEmail", ctx, "some@email.com").Return(nil, nil).Once()
		storage.On("GetByUsername", ctx, "some-username").Return(&User{}, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Email:    "some@email.com",
			Password: "some-password",
		})
		assert.ErrorIs(t, err, ErrUsernameTaken)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.Nil(t, res)
	})

	t.Run("Create with a database error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByEmail", ctx, "some@email.com").Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Email:    "some@email.com",
			Password: "some-password",
		})
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		user := User{
			ID:        uuid.UUID("some-id"),
			Username:  "some-username",
			Email:     "some-email",
			CreatedAt: time.Now(),
			password:  "some-encrypted-password",
		}

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		user := User{
			ID:        uuid.UUID("some-id"),
			Username:  "some-username",
			Email:     "some-email",
			CreatedAt: time.Now(),
			password:  "some-encrypted-password",
		}

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-invalid-password").Return(password.ErrMissmatchedPassword).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, "some-username", "some-invalid-password")
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		user := User{
			ID:        uuid.UUID("some-id"),
			Username:  "some-username",
			Email:     "some-email",
			CreatedAt: time.Now(),
			password:  "some-invalid-password",
		}

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-invalid-password", "some-password").Return(fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.EqualError(t, err, "failed to compare the hash and the password: some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(tools, storage)

		storage.On("GetByID", ctx, user.ID).Return(&user, nil).Once()

		res, err := service.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})
}
