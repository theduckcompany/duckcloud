package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_Users_Service(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		storage := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, storage, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()
		now := time.Now()
		newUser := User{
			id:                uuid.UUID("some-user-id"),
			username:          "Donald-Duck",
			createdAt:         now,
			passwordChangedAt: now,
			password:          secret.NewText("some-encrypted-password"),
			createdBy:         user.id,
			status:            Initializing,
			isAdmin:           false,
		}

		// Mocks
		storage.On("GetByUsername", ctx, "Donald-Duck").Return(nil, errNotFound).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-user-id")).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.PasswordMock.On("Encrypt", ctx, secret.NewText("my-super-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		storage.On("Save", ctx, &newUser).Return(nil)
		schedulerMock.On("RegisterUserCreateTask", mock.Anything, &scheduler.UserCreateArgs{UserID: uuid.UUID("some-user-id")}).
			Return(nil).Once()

		// Run
		res, err := service.Create(ctx, &CreateCmd{
			CreatedBy: user,
			Username:  "Donald-Duck",
			Password:  secret.NewText("my-super-password"),
			IsAdmin:   false,
		})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, &newUser, res)
	})

	t.Run("Create with a taken username", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(&User{}, nil).Once()

		// Run
		res, err := service.Create(ctx, &CreateCmd{
			CreatedBy: user,
			Username:  "Donald-Duck",
			Password:  secret.NewText("some-password"),
			IsAdmin:   false,
		})

		// Asserts
		require.ErrorIs(t, err, ErrUsernameTaken)
		require.ErrorIs(t, err, errs.ErrBadRequest)
		assert.Nil(t, res)
	})

	t.Run("Create with a database error", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(nil, fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			CreatedBy: user,
			Username:  "Donald-Duck",
			Password:  secret.NewText("some-secret"),
			IsAdmin:   false,
		})

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("Authenticate success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, user.password, secret.NewText("some-password")).Return(true, nil).Once()

		// Run
		res, err := service.Authenticate(ctx, "Donald-Duck", secret.NewText("some-password"))

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, user, res)
	})

	t.Run("Authenticate with an invalid username", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(nil, errNotFound).Once()

		// Run
		res, err := service.Authenticate(ctx, "Donald-Duck", secret.NewText("some-secret"))

		// Asserts
		require.ErrorIs(t, err, errs.ErrBadRequest)
		require.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, res)
	})

	t.Run("Authenticate with an invalid password", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, user.password, secret.NewText("some-invalid-password")).Return(false, nil).Once()

		// Invalid password here
		res, err := service.Authenticate(ctx, "Donald-Duck", secret.NewText("some-invalid-password"))
		require.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, res)
	})

	t.Run("Authenticate an unhandled password error", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByUsername", ctx, "Donald-Duck").Return(user, nil).Once()
		tools.PasswordMock.On("Compare", ctx, user.password, secret.NewText("some-password")).Return(false, fmt.Errorf("some-error")).Once()

		// Run
		res, err := service.Authenticate(ctx, "Donald-Duck", secret.NewText("some-password"))

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByID", ctx, user.ID()).Return(user, nil).Once()

		// Run
		res, err := service.GetByID(ctx, user.ID())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, user, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetAll", ctx, &sqlstorage.PaginateCmd{Limit: 10}).Return([]User{*user}, nil).Once()

		// Run
		res, err := service.GetAll(ctx, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []User{*user}, res)
	})

	t.Run("GetAllWithStatus success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetAll", ctx, &sqlstorage.PaginateCmd{Limit: 10}).Return([]User{*user}, nil).Once()

		// Run
		res, err := service.GetAllWithStatus(ctx, Active, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []User{*user}, res)
	})

	t.Run("AddToDeletion an admin user success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).WithAdminRole().Build()
		anAnotherAdmin := NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		store.On("GetByID", ctx, user.ID()).Return(user, nil).Once()
		store.On("GetAll", ctx, (*sqlstorage.PaginateCmd)(nil)).
			Return([]User{*user, *anAnotherAdmin}, nil).Once() // We check that the deleted user is not the last admin.
		schedulerMock.On("RegisterUserDeleteTask", mock.Anything, &scheduler.UserDeleteArgs{UserID: user.ID()}).
			Return(nil).Once()
		store.On("Patch", mock.Anything, user.ID(), map[string]any{"status": Deleting}).Return(nil).Once()

		// Run
		err := service.AddToDeletion(ctx, user.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("AddToDeletion with a user not found", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByID", ctx, user.ID()).Return(nil, errNotFound).Once()

		// Run
		err := service.AddToDeletion(ctx, user.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrNotFound)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("AddToDeletion the last admin failed", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).WithAdminRole().Build()
		anAnotherUser := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByID", ctx, user.ID()).Return(user, nil).Once()
		store.On("GetAll", ctx, (*sqlstorage.PaginateCmd)(nil)).Return([]User{*user, *anAnotherUser}, nil).Once() // This is the last admin

		err := service.AddToDeletion(ctx, user.ID())
		require.EqualError(t, err, "unauthorized: can't remove the last admin")
	})

	t.Run("HardDelete success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		someSoftDeletedUser := NewFakeUser(t).WithStatus(Deleting).Build()

		// Mocks
		store.On("GetByID", mock.Anything, someSoftDeletedUser.ID()).Return(someSoftDeletedUser, nil).Once()
		store.On("HardDelete", mock.Anything, someSoftDeletedUser.ID()).Return(nil).Once()

		// Run
		err := service.HardDelete(ctx, someSoftDeletedUser.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("HardDelete an non existing user", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		someSoftDeletedUser := NewFakeUser(t).WithStatus(Deleting).Build()

		// Mocks
		store.On("GetByID", mock.Anything, someSoftDeletedUser.ID()).Return(nil, errNotFound).Once()

		// Run
		err := service.HardDelete(ctx, someSoftDeletedUser.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("HardDelete an invalid status", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		someStillActifUser := NewFakeUser(t).WithStatus(Active).Build()

		// Mocks
		store.On("GetByID", mock.Anything, someStillActifUser.ID()).Return(someStillActifUser, nil).Once()

		// Run
		err := service.HardDelete(ctx, someStillActifUser.ID())

		// Asserts
		require.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("MarkInitAsFinished success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		someInitializingUser := NewFakeUser(t).WithStatus(Initializing).Build()

		// Mocks
		store.On("GetByID", mock.Anything, someInitializingUser.ID()).Return(someInitializingUser, nil).Once()
		store.On("Patch", mock.Anything, someInitializingUser.ID(), map[string]any{"status": Active}).Return(nil).Once()

		// Run
		res, err := service.MarkInitAsFinished(ctx, someInitializingUser.ID())

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, someInitializingUser, res)
	})

	t.Run("MarkInitAsFinished with a user with an invalid status", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		someAlreadyActifUser := NewFakeUser(t).WithStatus(Active).Build()

		// Mocks
		store.On("GetByID", mock.Anything, someAlreadyActifUser.ID()).Return(someAlreadyActifUser, nil).Once()

		// Run
		res, err := service.MarkInitAsFinished(ctx, someAlreadyActifUser.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("UpdatePassword success", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()
		now := time.Now()

		// Mocks
		store.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-new-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		store.On("Patch", mock.Anything, user.ID(), map[string]any{
			"password":            secret.NewText("some-encrypted-password"),
			"password_changed_at": sqlstorage.SQLTime(now),
		}).Return(nil).Once()

		// Run
		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      user.ID(),
			NewPassword: secret.NewText("some-new-password"),
		})

		// Asserts
		require.NoError(t, err)
	})

	t.Run("UpdatePassword with a user not found", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()

		// Mocks
		store.On("GetByID", mock.Anything, user.ID()).
			Return(nil, errs.ErrNotFound).Once()

		// Run
		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      user.ID(),
			NewPassword: secret.NewText("some-password"),
		})

		// Asserts
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("UpdatePassword with a patch error", func(t *testing.T) {
		t.Parallel()
		tools := tools.NewMock(t)
		store := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		service := newService(tools, store, schedulerMock)

		// Data
		user := NewFakeUser(t).Build()
		now := time.Now()

		// Mocks
		store.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		tools.PasswordMock.On("Encrypt", mock.Anything, secret.NewText("some-new-password")).
			Return(secret.NewText("some-encrypted-password"), nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		store.On("Patch", mock.Anything, user.ID(), map[string]any{
			"password":            secret.NewText("some-encrypted-password"),
			"password_changed_at": sqlstorage.SQLTime(now),
		}).Return(fmt.Errorf("some-error")).Once()

		// Run
		err := service.UpdateUserPassword(ctx, &UpdatePasswordCmd{
			UserID:      user.ID(),
			NewPassword: secret.NewText("some-new-password"),
		})

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
