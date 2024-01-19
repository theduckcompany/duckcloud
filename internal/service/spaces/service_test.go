package spaces

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_SpaceService(t *testing.T) {
	ctx := context.Background()

	// This AliceID is hardcoded in order to avoid dependency cycles
	const AliceID = uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")
	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalSpace.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalSpace).Return(nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			User:   &users.ExampleAlice,
			Name:   ExampleAlicePersonalSpace.name,
			Owners: []uuid.UUID{AliceID},
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		res, err := svc.Create(ctx, &CreateCmd{
			User:   &users.ExampleAlice,
			Name:   "",
			Owners: []uuid.UUID{AliceID},
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Name: cannot be blank.")
	})

	t.Run("Create with a non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		res, err := svc.Create(ctx, &CreateCmd{
			User:   &users.ExampleBob,
			Name:   "First space",
			Owners: []uuid.UUID{AliceID},
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("Create with a Save error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalSpace.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalSpace).Return(fmt.Errorf("some-error")).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			User:   &users.ExampleAlice,
			Name:   "Alice's Space",
			Owners: []uuid.UUID{AliceID},
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetAlluserSpaces success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllUserSpaces", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return([]Space{ExampleAlicePersonalSpace}, nil).Once()

		res, err := svc.GetAllUserSpaces(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Space{ExampleAlicePersonalSpace}, res)
	})

	t.Run("GetAlluserSpaces with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllUserSpaces", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetAllUserSpaces(ctx, AliceID, nil)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetByID with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(fmt.Errorf("some-error"))

		err := svc.Delete(ctx, ExampleAlicePersonalSpace.ID())
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("GetUserSpace not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetUserSpace with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace with an existing space but an invalid user id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetUserSpace(ctx, uuid.UUID("some-invalid-user-id"), ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
		assert.ErrorIs(t, err, ErrInvalidSpaceAccess)
	})

	t.Run("GetAllSpaces success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		storageMock.On("GetAllSpaces", mock.Anything, &storage.PaginateCmd{}).
			Return([]Space{ExampleAlicePersonalSpace, ExampleBobPersonalSpace}, nil).Once()

		res, err := svc.GetAllSpaces(ctx, &users.ExampleAlice, &storage.PaginateCmd{})
		assert.NoError(t, err)
		assert.Equal(t, res, []Space{ExampleAlicePersonalSpace, ExampleBobPersonalSpace})
	})

	t.Run("GetAllSpaces with a user not admin", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.False(t, users.ExampleBob.IsAdmin())

		res, err := svc.GetAllSpaces(ctx, &users.ExampleBob, &storage.PaginateCmd{StartAfter: map[string]string{}, Limit: 4})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("RemoveOwner success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalSpace.ID(), map[string]interface{}{
			"owners": Owners{},
		}).Return(nil).Once()

		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleAlice,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.NoError(t, err)

		expected := ExampleAlicePersonalSpace
		expected.owners = Owners{}
		assert.Equal(t, &expected, res)
	})

	t.Run("RemoveOwner with a non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.False(t, users.ExampleBob.IsAdmin())

		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    &users.ExampleBob,
			Owner:   &users.ExampleBob,
			SpaceID: ExampleBobPersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("RemoveOwner with a GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(nil, errs.ErrInternal).Once()

		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleAlice,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("RemoveOwner with a user not present in perms", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleBob, // Bob is not set as owner inside ExampleAlicePersonalSpace
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlicePersonalSpace, res) // nothing change
	})

	t.Run("RemoveOwner with a Patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalSpace.ID(), map[string]interface{}{
			"owners": Owners{},
		}).Return(errs.ErrInternal).Once()

		res, err := svc.RemoveOwner(ctx, &RemoveOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleAlice,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("AddOwner success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalSpace.ID(), map[string]interface{}{
			"owners": Owners{users.ExampleAlice.ID(), users.ExampleBob.ID()},
		}).Return(nil).Once()

		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleBob,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.NoError(t, err)

		expected := ExampleAlicePersonalSpace
		expected.owners = Owners{users.ExampleAlice.ID(), users.ExampleBob.ID()}
		assert.Equal(t, &expected, res)
	})

	t.Run("AddOwner with a User not admin", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.False(t, users.ExampleBob.IsAdmin())

		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    &users.ExampleBob,
			Owner:   &users.ExampleBob,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("AddOwner with a GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(nil, errs.ErrInternal).Once()

		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleBob,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("AddOwner with a user already present in perms", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleAlice, // Alice is already present in perms
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlicePersonalSpace, res) // nothing change
	})

	t.Run("AddOwner with a Patch error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		require.True(t, users.ExampleAlice.IsAdmin())

		// copy the struct to avoid any changes and impact on other tests
		copyAliceSpace := ExampleAlicePersonalSpace

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).
			Return(&copyAliceSpace, nil).Once()

		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalSpace.ID(), map[string]interface{}{
			"owners": Owners{users.ExampleAlice.ID(), users.ExampleBob.ID()},
		}).Return(errs.ErrInternal).Once()

		res, err := svc.AddOwner(ctx, &AddOwnerCmd{
			User:    &users.ExampleAlice,
			Owner:   &users.ExampleBob,
			SpaceID: ExampleAlicePersonalSpace.ID(),
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("Bootstrap success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllSpaces", mock.Anything, &storage.PaginateCmd{Limit: 1}).
			Return([]Space{}, nil).Once()

		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   BootstrapSpaceName,
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		}).Return(nil).Once()

		err := svc.Bootstrap(ctx, &users.ExampleAlice)
		assert.NoError(t, err)
	})

	t.Run("Bootstrap with a GetAllSpaces error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllSpaces", mock.Anything, &storage.PaginateCmd{Limit: 1}).
			Return(nil, errs.ErrInternal).Once()

		err := svc.Bootstrap(ctx, &users.ExampleAlice)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("Bootstrap with an already bootstraped service", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllSpaces", mock.Anything, &storage.PaginateCmd{Limit: 1}).
			Return([]Space{ExampleAlicePersonalSpace}, nil).Once()

		err := svc.Bootstrap(ctx, &users.ExampleAlice)
		assert.NoError(t, err)
	})

	t.Run("Bootstrap with a Scheduler error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewService(tools, storageMock, schedulerMock)

		storageMock.On("GetAllSpaces", mock.Anything, &storage.PaginateCmd{Limit: 1}).
			Return([]Space{}, nil).Once()

		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   BootstrapSpaceName,
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		}).Return(errs.ErrInternal).Once()

		err := svc.Bootstrap(ctx, &users.ExampleAlice)
		assert.ErrorIs(t, err, errs.ErrInternal)
	})
}
