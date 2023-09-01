package folders

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_FolderService(t *testing.T) {
	ctx := context.Background()

	// This AliceID is hardcoded in order to avoid dependency cycles
	const AliceID = uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	t.Run("CreatePersonalFolder success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		inodesMock.On("CreateRootDir", mock.Anything).Return(&inodes.ExampleAliceRoot, nil).Once()
		tools.UUIDMock.On("New").Return(ExampleAlicePersonalFolder.ID()).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAlicePersonalFolder).Return(nil).Once()

		res, err := svc.CreatePersonalFolder(ctx, &CreatePersonalFolderCmd{
			Name:  "Alice's Folder",
			Owner: AliceID,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("CreatePersonalFolder with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		res, err := svc.CreatePersonalFolder(ctx, &CreatePersonalFolderCmd{
			Name:  "",
			Owner: AliceID,
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: Name: cannot be blank.")
	})

	t.Run("GetAlluserFolders success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetAllUserFolders", mock.Anything, AliceID, (*storage.PaginateCmd)(nil)).Return([]Folder{ExampleAlicePersonalFolder}, nil).Once()

		res, err := svc.GetAllUserFolders(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Folder{ExampleAlicePersonalFolder}, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()

		res, err := svc.GetByID(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.RootFS()).Return(nil, nil).Once()
		storageMock.On("Delete", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete an non existing folder", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(nil, nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
	})

	t.Run("Delete with a root inode still present", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.ID()).Return(&ExampleAlicePersonalFolder, nil).Once()
		inodesMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.RootFS()).Return(&inodes.ExampleAliceRoot, nil).Once()

		err := svc.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.ErrorIs(t, err, ErrRootFSExist)
	})
}
