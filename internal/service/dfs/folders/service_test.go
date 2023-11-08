package folders

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

func Test_FolderService(t *testing.T) {
	ctx := context.Background()

	// This AliceID is hardcoded in order to avoid dependency cycles
	const AliceID = uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalFolder.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalFolder).Return(nil).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   ExampleAlicePersonalFolder.name,
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalFolder.rootFS,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   "",
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalFolder.rootFS,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "Name: cannot be blank.")
	})

	t.Run("Create with a Save error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		tools.UUIDMock.On("New").Return(ExampleAlicePersonalFolder.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalFolder).Return(fmt.Errorf("some-error")).Once()

		res, err := svc.Create(ctx, &CreateCmd{
			Name:   "Alice's Folder",
			Owners: []uuid.UUID{AliceID},
			RootFS: ExampleAlicePersonalFolder.rootFS,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetAlluserFolders success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetAllUserFolders", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return([]Folder{ExampleAlicePersonalFolder}, nil).Once()

		res, err := svc.GetAllUserFolders(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Folder{ExampleAlicePersonalFolder}, res)
	})

	t.Run("GetAlluserFolders with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetAllUserFolders", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetAllUserFolders(ctx, AliceID, nil)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalFolder.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetByID with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalFolder.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(fmt.Errorf("some-error"))

		err := svc.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserFolder success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()

		res, err := svc.GetUserFolder(ctx, ExampleAlicePersonalFolder.Owners()[0], ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("GetUserFolder not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil, errNotFound).Once()

		res, err := svc.GetUserFolder(ctx, ExampleAlicePersonalFolder.Owners()[0], ExampleAlicePersonalFolder.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("GetUserFolder with an error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil, fmt.Errorf("some-error")).Once()

		res, err := svc.GetUserFolder(ctx, ExampleAlicePersonalFolder.Owners()[0], ExampleAlicePersonalFolder.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("GetUserFolder with an existing folder but an invalid user id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()

		res, err := svc.GetUserFolder(ctx, uuid.UUID("some-invalid-user-id"), ExampleAlicePersonalFolder.ID())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
		assert.ErrorIs(t, err, ErrInvalidFolderAccess)
	})
}
