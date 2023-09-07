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

	t.Run("RegisterWrite success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.id).Return(&ExampleAlicePersonalFolder, nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalFolder.id, map[string]any{
			"last_modified_at": now,
			"size":             uint64(42),
		}).Return(nil).Once()

		expected := ExampleAlicePersonalFolder
		expected.lastModifiedAt = now
		expected.size = uint64(42)

		res, err := svc.RegisterWrite(ctx, ExampleAlicePersonalFolder.id, 42)
		assert.NoError(t, err)
		assert.Equal(t, &expected, res)
	})

	t.Run("RegisterWrite with an invalid id", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("some-invalid-id")).Return(nil, nil).Once()

		res, err := svc.RegisterWrite(ctx, "some-invalid-id", 42)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("GetAllFoldersWithRoot success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		storageMock.On("GetAllFoldersWithRoot", mock.Anything, ExampleAlicePersonalFolder.rootFS, (*storage.PaginateCmd)(nil)).Return([]Folder{ExampleAlicePersonalFolder}, nil).Once()

		res, err := svc.GetAllFoldersWithRoot(ctx, ExampleAlicePersonalFolder.rootFS, nil)
		assert.NoError(t, err)
		assert.Equal(t, []Folder{ExampleAlicePersonalFolder}, res)
	})

	t.Run("RegisterDeletion success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		folderWithContent := ExampleAlicePersonalFolder
		folderWithContent.size = uint64(42)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.id).Return(&folderWithContent, nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalFolder.id, map[string]any{
			"last_modified_at": now,
			"size":             uint64(32),
		}).Return(nil).Once()

		expected := ExampleAlicePersonalFolder
		expected.lastModifiedAt = now
		expected.size = uint64(32)

		// Register a file deletion of size 10
		res, err := svc.RegisterDeletion(ctx, ExampleAlicePersonalFolder.id, 10)
		assert.NoError(t, err)
		assert.Equal(t, &expected, res)
	})

	t.Run("RegisterDeletion with a biggest size than the folder size", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewService(tools, storageMock, inodesMock)

		folderWithContent := ExampleAlicePersonalFolder
		folderWithContent.size = uint64(42)

		storageMock.On("GetByID", mock.Anything, ExampleAlicePersonalFolder.id).Return(&folderWithContent, nil).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAlicePersonalFolder.id, map[string]any{
			"last_modified_at": now,
			"size":             uint64(0),
		}).Return(nil).Once()

		expected := ExampleAlicePersonalFolder
		expected.lastModifiedAt = now
		expected.size = uint64(0)

		// Register a file deletion of size 120 which is bigger than 42
		res, err := svc.RegisterDeletion(ctx, ExampleAlicePersonalFolder.id, 120)
		assert.NoError(t, err)
		assert.Equal(t, &expected, res)
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
