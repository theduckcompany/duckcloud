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

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, "some-password").Return(ExampleAlice.password, nil).Once()

		store.On("Save", ctx, &ExampleInitializingAlice).Return(nil)

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: "some-password",
			IsAdmin:  true,
		})
		assert.NoError(t, err)

		assert.Equal(t, &ExampleInitializingAlice, res)
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

	t.Run("GetAllWithStatus success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleAlice}, nil).Once()

		res, err := service.GetAllWithStatus(ctx, "active", &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("AddToDeletion success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		anAnotherAdmin := ExampleBob
		anAnotherAdmin.isAdmin = true

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice, anAnotherAdmin}, nil).Once()
		store.On("Patch", ctx, ExampleAlice.ID(), map[string]any{"status": "deleting"}).Return(nil).Once()

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("AddToDeletion with a user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(nil, nil).Once()

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("AddToDeletion the last admin failed", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice}, nil).Once() // This is the last admin

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.EqualError(t, err, "unauthorized: can't remove the last admin")
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleDeletingAlice, nil).Once()
		store.On("HardDelete", mock.Anything, ExampleAlice.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an non existing user", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		// It doesn't return ExampleDeletingAlice so the status is "active"
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(nil, nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodes := inodes.NewMockService(t)
		service := NewService(tools, store, inodes)

		// It doesn't return ExampleDeletingAlice so the status is "active"
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("SaveBootstrapInfos success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(tools, store, inodesMock)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleInitializingAlice, nil).Once()
		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]any{
			"fs_root": inodes.ExampleAliceRoot.ID(),
			"status":  "active",
		}).Return(nil).Once()

		res, err := service.SaveBootstrapInfos(ctx, ExampleAlice.ID(), &inodes.ExampleAliceRoot)
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlice, res)
	})

	t.Run("SaveBootstrapInfos with a user with an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		inodesMock := inodes.NewMockService(t)
		service := NewService(tools, store, inodesMock)

		// ExampleAlice is already initialized.
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		res, err := service.SaveBootstrapInfos(ctx, ExampleAlice.ID(), &inodes.ExampleAliceRoot)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})
}
