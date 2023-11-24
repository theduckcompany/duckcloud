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

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(nil, errNotFound).Once()

		tools.UUIDMock.On("New").Return(ExampleBob.ID()).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, secret.NewText("some-password")).Return(ExampleBob.password, nil).Once()

		store.On("Save", ctx, &ExampleInitializingBob).Return(nil)
		schedulerMock.On("RegisterUserCreateTask", mock.Anything, &scheduler.UserCreateArgs{UserID: ExampleBob.ID()}).
			Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			User:     &ExampleAlice,
			Username: ExampleBob.Username(),
			Password: secret.NewText("some-password"),
			IsAdmin:  false,
		})
		assert.NoError(t, err)

		assert.Equal(t, &ExampleInitializingBob, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(&User{}, nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			User:     &ExampleAlice,
			Username: ExampleBob.Username(),
			Password: ExampleBob.password,
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

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			User:     &ExampleAlice,
			Username: ExampleBob.Username(),
			Password: ExampleBob.password,
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

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(&ExampleBob, nil).Once()

		tools.PasswordMock.On("Compare", ctx, ExampleBob.password, secret.NewText("some-password")).Return(true, nil).Once()

		res, err := service.Authenticate(ctx, ExampleBob.Username(), secret.NewText("some-password"))
		assert.NoError(t, err)
		assert.Equal(t, &ExampleBob, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(nil, errNotFound).Once()

		res, err := service.Authenticate(ctx, ExampleBob.Username(), ExampleBob.password)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(&ExampleBob, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleBob.password, secret.NewText("some-invalid-password")).Return(false, nil).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, ExampleBob.Username(), secret.NewText("some-invalid-password"))
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByUsername", ctx, ExampleBob.Username()).Return(&ExampleBob, nil).Once()
		tools.PasswordMock.On("Compare", ctx, ExampleBob.password, secret.NewText("some-password")).Return(false, fmt.Errorf("some-error")).Once()

		res, err := service.Authenticate(ctx, ExampleBob.Username(), secret.NewText("some-password"))
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", ctx, ExampleBob.ID()).Return(&ExampleBob, nil).Once()

		res, err := service.GetByID(ctx, ExampleBob.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleBob, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleBob}, nil).Once()

		res, err := service.GetAll(ctx, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleBob}, res)
	})

	t.Run("GetAllWithStatus success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetAll", ctx, &storage.PaginateCmd{Limit: 10}).Return([]User{ExampleBob}, nil).Once()

		res, err := service.GetAllWithStatus(ctx, Active, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleBob}, res)
	})

	t.Run("AddToDeletion success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		anAnotherAdmin := ExampleAlice
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

		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&ExampleDeletingAlice, nil).Once()
		store.On("HardDelete", mock.Anything, ExampleBob.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleBob.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an non existing user", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// It doesn't return ExampleDeletingBob so the status is Active
		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(nil, errNotFound).Once()

		err := service.HardDelete(ctx, ExampleBob.ID())
		assert.NoError(t, err)
	})

	t.Run("HardDelete an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// It doesn't return ExampleDeletingBob so the status is Active
		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&ExampleBob, nil).Once()

		err := service.HardDelete(ctx, ExampleBob.ID())
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("SetDefaultSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("Patch", mock.Anything, ExampleBob.ID(), map[string]interface{}{
			"space": spaces.ExampleAliceBobSharedSpace.ID(),
		}).Return(nil).Once()

		res, err := service.SetDefaultSpace(ctx, ExampleBob, &spaces.ExampleAliceBobSharedSpace)
		assert.NoError(t, err)
		expected := ExampleBob
		expected.defaultSpaceID = spaces.ExampleAliceBobSharedSpace.ID()
		assert.Equal(t, &expected, res)
	})

	t.Run("SetDefaultSpace with a space not owned by the user", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// BobPersonalSpace is not owned by Bob
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

		initializingBob := ExampleInitializingBob
		initializingBob.defaultSpaceID = spaces.ExampleBobPersonalSpace.ID()

		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&initializingBob, nil).Once()
		store.On("Patch", mock.Anything, ExampleBob.ID(), map[string]any{"status": Active}).Return(nil).Once()

		res, err := service.MarkInitAsFinished(ctx, ExampleBob.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleBob, res)
	})

	t.Run("MarkInitAsFinished with a user with an invalid status", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		// ExampleBob is already initialized.
		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&ExampleBob, nil).Once()

		res, err := service.MarkInitAsFinished(ctx, ExampleBob.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("UpdatePassword success", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&ExampleBob, nil).Once()

		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		store.On("Patch", mock.Anything, ExampleBob.ID(), map[string]any{
			"password": secret.NewText("some-encrypted-password"),
		}).Return(nil).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleBob.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.NoError(t, err)
	})

	t.Run("UpdatePassword with a user not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleBob.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleBob.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("UpdatePassword with a patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		store := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := NewService(tools, store, schedulerMock)

		store.On("GetByID", mock.Anything, ExampleBob.ID()).Return(&ExampleBob, nil).Once()

		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		store.On("Patch", mock.Anything, ExampleBob.ID(), map[string]any{
			"password": secret.NewText("some-encrypted-password"),
		}).Return(fmt.Errorf("some-error")).Once()

		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      ExampleBob.ID(),
			NewPassword: secret.NewText("some-password"),
		})
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
