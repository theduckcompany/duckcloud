package spaces

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_SpaceService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		now := time.Now()
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).
			CreatedBy(user).
			CreatedAt(now).
			WithOwners(*user).
			WithName("Donald's space").
			Build()

		// Mocks
		tools.UUIDMock.On("New").Return(someSpace.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, someSpace).Return(nil).Once()

		// Run
		res, err := svc.Create(ctx, &CreateCmd{
			User:   user,
			Name:   "Donald's space",
			Owners: []uuid.UUID{user.ID()},
		})

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, someSpace, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Run
		res, err := svc.Create(ctx, &CreateCmd{
			User:   user,
			Name:   "",
			Owners: []uuid.UUID{},
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "Name: cannot be blank.")
	})

	t.Run("Create with a non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		notAnAdminUser := users.NewFakeUser(t).Build()

		res, err := svc.Create(ctx, &CreateCmd{
			User:   notAnAdminUser,
			Name:   "Donald's space",
			Owners: []uuid.UUID{},
		})
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("Create with a Save error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		now := time.Now()
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).
			CreatedBy(user).
			CreatedAt(now).
			WithName("Some space").
			Build()

		// Mocks
		tools.UUIDMock.On("New").Return(someSpace.id).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, someSpace).Return(fmt.Errorf("some-error")).Once()

		// Run
		res, err := svc.Create(ctx, &CreateCmd{
			User:   user,
			Name:   "Some space",
			Owners: []uuid.UUID{},
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("GetAlluserSpaces success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]Space{*someSpace}, nil).Once()

		// Run
		res, err := svc.GetAllUserSpaces(ctx, user.ID(), nil)

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, []Space{*someSpace}, res)
	})

	t.Run("GetAlluserSpaces with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		// Run
		res, err := svc.GetAllUserSpaces(ctx, user.ID(), nil)

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		// Run
		res, err := svc.GetByID(ctx, someSpace.ID())

		// Assert
		require.NoError(t, err)
		assert.EqualValues(t, someSpace, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, errNotFound).Once()

		// Run
		res, err := svc.GetByID(ctx, someSpace.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetByID with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		// Run
		res, err := svc.GetByID(ctx, someSpace.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someSpace := NewFakeSpace(t).Build()
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("Delete", mock.Anything, someSpace.ID()).Return(nil).Once()

		// Run
		err := svc.Delete(ctx, user, someSpace.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Delete with an non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		nontAdminUser := users.NewFakeUser(t).Build() // Not an admin
		someSpace := NewFakeSpace(t).Build()

		// Mocks

		// Run
		err := svc.Delete(ctx, nontAdminUser, someSpace.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("Delete with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("Delete", mock.Anything, someSpace.ID()).Return(fmt.Errorf("some-error"))

		// Run
		err := svc.Delete(ctx, user, someSpace.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		// Run
		res, err := svc.GetUserSpace(ctx, user.ID(), someSpace.ID())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, someSpace, res)
	})

	t.Run("GetUserSpace not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, errNotFound).Once()

		// Run
		res, err := svc.GetUserSpace(ctx, user.ID(), someSpace.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetUserSpace with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		// Run
		res, err := svc.GetUserSpace(ctx, user.ID(), someSpace.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace with an existing space but an invalid user id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		// Run
		res, err := svc.GetUserSpace(ctx, uuid.UUID("some-invalid-user-id"), someSpace.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrUnauthorized)
		require.ErrorIs(t, err, ErrInvalidSpaceAccess)
	})

	t.Run("GetAllSpaces success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace1 := NewFakeSpace(t).WithOwners(*user).Build()
		someSpace2 := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetAllSpaces", mock.Anything, &sqlstorage.PaginateCmd{}).
			Return([]Space{*someSpace1, *someSpace2}, nil).Once()

		// Run
		res, err := svc.GetAllSpaces(ctx, user, &sqlstorage.PaginateCmd{})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Space{*someSpace1, *someSpace2}, res)
	})

	t.Run("GetAllSpaces with a user not admin", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someNonAdminUser := users.NewFakeUser(t).Build()

		// Run
		res, err := svc.GetAllSpaces(ctx, someNonAdminUser, &sqlstorage.PaginateCmd{StartAfter: map[string]string{}, Limit: 4})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("RemoveOwner success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()
		storageMock.On("Patch", mock.Anything, someSpace.ID(), map[string]interface{}{
			"owners": Owners{},
		}).Return(nil).Once()

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    user,
			Owner:   user,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		require.NoError(t, err)

		assert.Empty(t, res.owners)
	})

	t.Run("RemoveOwner with a non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someNonAdminUser := users.NewFakeUser(t).Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*someNonAdminUser).Build()

		// Mocks

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    someNonAdminUser,
			Owner:   someOtherUser,
			SpaceID: someSpace.id,
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("RemoveOwner with a non admin user removing itself", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someNonAdminUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*someNonAdminUser).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).
			Return(someSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, someSpace.ID(), map[string]interface{}{
			"owners": Owners{},
		}).Return(nil).Once()

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    someNonAdminUser,
			Owner:   someNonAdminUser,
			SpaceID: someSpace.id,
		})

		// Asserts
		require.NoError(t, err)
		assert.Empty(t, res.owners)
	})

	t.Run("RemoveOwner with a GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someAdminUser := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*someAdminUser).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, errs.ErrInternal).Once()

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    someAdminUser,
			Owner:   someAdminUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("RemoveOwner with a user not present in perms", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someAdminUser := users.NewFakeUser(t).WithAdminRole().Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*someOtherUser).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    someAdminUser,
			Owner:   someAdminUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, Owners{someOtherUser.ID()}, res.owners) // nothing change
	})

	t.Run("RemoveOwner with a Patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, someSpace.ID(), map[string]interface{}{"owners": Owners{}}).
			Return(errs.ErrInternal).Once()

		// Run
		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    user,
			Owner:   user,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("AddOwner success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*user).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()
		storageMock.On("Patch", mock.Anything, someSpace.ID(), map[string]interface{}{
			"owners": Owners{user.ID(), someOtherUser.ID()},
		}).Return(nil).Once()

		// Run
		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    user,
			Owner:   someOtherUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, Owners{user.ID(), someOtherUser.ID()}, res.owners)
	})

	t.Run("AddOwner with a User not admin", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		someNonAdminUser := users.NewFakeUser(t).Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).Build()

		// Mocks

		// Run
		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    someNonAdminUser,
			Owner:   someOtherUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("AddOwner with a GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(nil, errs.ErrInternal).Once()

		// Run
		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    user,
			Owner:   someOtherUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("AddOwner with a user already present in perms", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*someOtherUser).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		// Run
		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    user,
			Owner:   someOtherUser,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, Owners{someOtherUser.ID()}, res.owners) // nothing change
	})

	t.Run("AddOwner with a Patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someOtherUser := users.NewFakeUser(t).Build()
		someSpace := NewFakeSpace(t).WithOwners(*someOtherUser).Build()

		// Mocks
		storageMock.On("GetByID", mock.Anything, someSpace.ID()).Return(someSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, someSpace.ID(), map[string]interface{}{
			"owners": Owners{someOtherUser.ID(), user.ID()},
		}).Return(errs.ErrInternal).Once()

		// Run
		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    user,
			Owner:   user,
			SpaceID: someSpace.ID(),
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("Bootstrap success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("GetAllSpaces", mock.Anything, &sqlstorage.PaginateCmd{Limit: 1}).
			Return([]Space{}, nil).Once()

		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: user.ID(),
			Name:   BootstrapSpaceName,
			Owners: []uuid.UUID{user.ID()},
		}).Return(nil).Once()

		// Run
		err := svc.Bootstrap(ctx, user)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Bootstrap with a GetAllSpaces error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("GetAllSpaces", mock.Anything, &sqlstorage.PaginateCmd{Limit: 1}).
			Return(nil, errs.ErrInternal).Once()

		// Run
		err := svc.Bootstrap(ctx, user)

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("Bootstrap with an already bootstraped service", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		someSpace := NewFakeSpace(t).Build()

		// Mocks
		storageMock.On("GetAllSpaces", mock.Anything, &sqlstorage.PaginateCmd{Limit: 1}).
			Return([]Space{*someSpace}, nil).Once()

		// Run
		err := svc.Bootstrap(ctx, user)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Bootstrap with a Scheduler error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := newService(tools, storageMock, schedulerMock)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()

		// Mocks
		storageMock.On("GetAllSpaces", mock.Anything, &sqlstorage.PaginateCmd{Limit: 1}).
			Return([]Space{}, nil).Once()

		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: user.ID(),
			Name:   BootstrapSpaceName,
			Owners: []uuid.UUID{user.ID()},
		}).Return(errs.ErrInternal).Once()

		// Run
		err := svc.Bootstrap(ctx, user)

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
	})
}
