package users

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/password"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func Test_Users_Service(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodesSvc := inodes.NewMockService(t)
		service := NewService(tools, store, inodesSvc)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, nil).Once()

		tools.UUIDMock.On("New").Return(ExampleAlice.ID()).Once()

		inodesSvc.On("CreateRootDir", ctx, ExampleAlice.ID()).Return(&inodes.ExampleAliceRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, "some-password").Return(ExampleAlice.password, nil).Once()

		store.On("Save", ctx, &ExampleAlice).Return(nil)

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: "some-password",
			IsAdmin:  true,
		})
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&User{}, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: ExampleAlice.password,
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

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: ExampleAlice.password,
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

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()

		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, "some-password").Return(nil).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), "some-password")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, nil).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), ExampleAlice.password)
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, "some-invalid-password").Return(password.ErrMissmatchedPassword).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, ExampleAlice.Username(), "some-invalid-password")
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, "some-password").Return(fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), "some-password")
		assert.EqualError(t, err, "failed to compare the hash and the password: some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		res, err := service.GetByID(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleAlice}, nil).Once()

		res, err := service.GetAll(ctx, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		anAnotherAdmin := ExampleBob
		anAnotherAdmin.isAdmin = true

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice, anAnotherAdmin}, nil).Once()
		store.On("Delete", ctx, ExampleAlice.ID()).Return(nil).Once()

		err := service.Delete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete with a user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(nil, nil).Once()

		err := service.Delete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete the last admin failed", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice}, nil).Once() // This is the last admin

		err := service.Delete(ctx, ExampleAlice.ID())
		assert.EqualError(t, err, "unauthorized: can't remove the last admin")
	})

	t.Run("GetAllDeleted success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetAllDeleted", mock.Anything, 10).Return([]User{ExampleAlice}, nil).Once()

		res, err := service.GetAllDeleted(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, ExampleAlice, res[0])
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetDeleted", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("HardDelete", mock.Anything, ExampleAlice.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an non sofdeleted inode does nothing", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetDeleted", mock.Anything, ExampleAlice.ID()).Return(nil, nil).Once()
		// The HardeDelete method is not called as we haven't found the deletedINode

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})
}
