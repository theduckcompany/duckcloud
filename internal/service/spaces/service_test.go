package spaces

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		svc := NewService(tools, storageMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalSpace.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalSpace).Return(nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   ExampleAlicePersonalSpace.name,
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalSpace.rootFS,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   "",
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalSpace.rootFS,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Name: cannot be blank.")
	})

	t.Run("Create with a Save error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalSpace.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalSpace).Return(fmt.Errorf("some-error")).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   "Alice's Space",
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalSpace.rootFS,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetAlluserSpaces success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetAllUserSpaces", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return([]Space{ExampleAlicePersonalSpace}, nil).Once()

		res, err := svc.GetAllUserSpaces(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Space{ExampleAlicePersonalSpace}, res)
	})

	t.Run("GetAlluserSpaces with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetAllUserSpaces", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetAllUserSpaces(ctx, AliceID, nil)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetByID with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(fmt.Errorf("some-error"))

		err := svc.Delete(ctx, ExampleAlicePersonalSpace.ID())
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("GetUserSpace not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetUserSpace with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetUserSpace(ctx, ExampleAlicePersonalSpace.Owners()[0], ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserSpace with an existing space but an invalid user id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalSpace.ID()).Return(&ExampleAlicePersonalSpace, nil).Once()

		res, err := svc.GetUserSpace(ctx, uuid.UUID("some-invalid-user-id"), ExampleAlicePersonalSpace.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
		assert.ErrorIs(t, err, ErrInvalidSpaceAccess)
	})
}
