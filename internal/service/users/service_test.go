package users

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func Test_Users_Service(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, errNotFound).Once()

		tools.UUIDMock.On("New").Return(ExampleAlice.ID()).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, secret.NewText("some-password")).Return(ExampleAlice.password, nil).Once()

		store.On("Save", ctx, &ExampleInitializingAlice).Return(nil)
		schedulerMock.On("RegisterUserCreateTask", mock.Anything, &scheduler.UserCreateArgs{UserID: ExampleAlice.ID()}).
			Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: secret.NewText("some-password"),
			IsAdmin:  true,
		})
		assert.NoError(t, err)

		assert.Equal(t, &ExampleInitializingAlice, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

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
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			Username: ExampleAlice.Username(),
			Password: ExampleAlice.password,
			IsAdmin:  false,
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()

		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, secret.NewText("some-password")).Return(true, nil).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), secret.NewText("some-password"))
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(nil, errNotFound).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), ExampleAlice.password)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, secret.NewText("some-invalid-password")).Return(false, nil).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, ExampleAlice.Username(), secret.NewText("some-invalid-password"))
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleAlice.Username()).Return(&ExampleAlice, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleAlice.password, secret.NewText("some-password")).Return(false, fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, ExampleAlice.Username(), secret.NewText("some-password"))
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		res, err := service.GetByID(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleAlice}, nil).Once()

		res, err := service.GetAll(ctx, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("GetAllWithStatus success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleAlice}, nil).Once()

		res, err := service.GetAllWithStatus(ctx, Active, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("AddToDeletion success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		anAnotherAdmin := ExampleBob
		anAnotherAdmin.isAdmin = true

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice, anAnotherAdmin}, nil).Once()
		schedulerMock.On("RegisterUserDeleteTask", mock.Anything, &scheduler.UserDeleteArgs{UserID: ExampleAlice.ID()}).
			Return(nil).Once()
		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]any{"status": Deleting}).Return(nil).Once()

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("AddToDeletion with a user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(nil, errNotFound).Once()

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("AddToDeletion the last admin failed", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", ctx, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()
		store.On("GetAll", ctx, (*storage.PaginateCmd)(nil)).Return([]User{ExampleAlice}, nil).Once() // This is the last admin

		err := service.AddToDeletion(ctx, ExampleAlice.ID())
		assert.EqualError(t, err, "unauthorized: can't remove the last admin")
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleDeletingAlice, nil).Once()
		store.On("HardDelete", mock.Anything, ExampleAlice.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an non existing user", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// It doesn't return ExampleDeletingAlice so the status is Active
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(nil, errNotFound).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// It doesn't return ExampleDeletingAlice so the status is Active
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		err := service.HardDelete(ctx, ExampleAlice.ID())
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("SetDefaultSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]interface{}{
			"space": spaces.ExampleAliceBobSharedSpace.ID(),
		}).Return(nil).Once()

		res, err := service.SetDefaultSpace(ctx, ExampleAlice, &spaces.ExampleAliceBobSharedSpace)
		assert.NoError(t, err)
		expected := ExampleAlice
		expected.defaultSpaceID = spaces.ExampleAliceBobSharedSpace.ID()
		assert.Equal(t, &expected, res)
	})

	t.Run("SetDefaultSpace with a space not owned by the user", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// BobPersonalSpace is not owned by Alice
		res, err := service.SetDefaultSpace(ctx, ExampleAlice, &spaces.ExampleBobPersonalSpace)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
		assert.ErrorIs(t, err, ErrUnauthorizedSpace)
		assert.Nil(t, res)
	})

	t.Run("MarkInitAsFinished success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		initializingAlice := ExampleInitializingAlice
		initializingAlice.defaultSpaceID = spaces.ExampleAlicePersonalSpace.ID()

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&initializingAlice, nil).Once()
		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]any{"status": Active}).Return(nil).Once()

		res, err := service.MarkInitAsFinished(ctx, ExampleAlice.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlice, res)
	})

	t.Run("MarkInitAsFinished with a user with an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// ExampleAlice is already initialized.
		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		res, err := service.MarkInitAsFinished(ctx, ExampleAlice.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("UpdatePassword success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]any{
			"password": secret.NewText("some-encrypted-password"),
		}).Return(nil).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleAlice.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.NoError(t, err)
	})

	t.Run("UpdatePassword with a user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleAlice.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("UpdatePassword with a patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		store.On("Patch", mock.Anything, ExampleAlice.ID(), map[string]any{
			"password": secret.NewText("some-encrypted-password"),
		}).Return(fmt.Errorf("some-error")).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleAlice.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
