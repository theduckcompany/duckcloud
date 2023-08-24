package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/password"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_Users_Service(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	user := User{
		id:        uuid.UUID("some-user-id"),
		username:  "some-username",
		isAdmin:   true,
		fsRoot:    uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
		password:  "some-encrypted-password",
		createdAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodesSvc := inodes.NewMockService(t)
		service := NewService(tools, store, inodesSvc)

		store.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-user-id")).Once()

		inodesSvc.On("BootstrapUser", ctx, uuid.UUID("some-user-id")).Return(&inodes.ExampleAliceRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, "some-password").Return("some-encrypted-password", nil).Once()

		store.On("Save", ctx, &user).Return(nil)

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Password: "some-password",
			IsAdmin:  true,
		})
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(&User{}, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Password: "some-password",
			IsAdmin:  false,
		})
		assert.ErrorIs(t, err, ErrUsernameTaken)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.Nil(t, res)
	})

	t.Run("Create with a database error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: "some-username",
			Password: "some-password",
			IsAdmin:  false,
		})
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()

		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(password.ErrMissmatchedPassword).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.EqualError(t, err, "failed to compare the hash and the password: some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, user.ID()).Return(&user, nil).Once()

		res, err := service.GetByID(ctx, user.ID())
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{user}, nil).Once()

		res, err := service.GetAll(ctx, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{user}, res)
	})
}
