package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/service/inodes"
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
		ID:        uuid.UUID("some-user-id"),
		Username:  "some-username",
		Email:     "some@email.com",
		FSRoot:    uuid.UUID("some-inode-id"),
		password:  "some-encrypted-password",
		CreatedAt: now,
	}

	inode := inodes.INode{
		ID:             uuid.UUID("some-inode-id"),
		UserID:         uuid.UUID("some-user-id"),
		Parent:         inodes.NoParent,
		Type:           inodes.Directory,
		CreatedAt:      now,
		LastModifiedAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByEmail", ctx, "some@email.com").Return(nil, nil).Once()
		storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-user-id")).Once()

		inodes.On("BootstrapUser", ctx, uuid.UUID("some-user-id")).Return(&inode, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, "some-password").Return("some-encrypted-password", nil).Once()

		storage.On("Save", ctx, &user).Return(nil)

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
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

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
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

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
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

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
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()

		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(password.ErrMissmatchedPassword).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, "some-encrypted-password", "some-password").Return(fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, "some-username", "some-password")
		assert.EqualError(t, err, "failed to compare the hash and the password: some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, storage, inodes)

		storage.On("GetByID", ctx, user.ID).Return(&user, nil).Once()

		res, err := service.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, &user, res)
	})
}
