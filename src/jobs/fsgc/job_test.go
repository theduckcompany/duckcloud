package fsgc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestFSGC(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		// Retrieve all the impacted folders
		inodesMock.On("GetINodeRoot", mock.Anything, &inodes.ExampleAliceFile).Return(&inodes.ExampleAliceRoot, nil).Once()
		foldersMock.On("GetAllFoldersWithRoot", mock.Anything, inodes.ExampleAliceRoot.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()
		foldersMock.On("RegisterDeletion", mock.Anything, folders.ExampleAlicePersonalFolder.ID(), uint64(inodes.ExampleAliceFile.Size())).Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		// We remove the file content and inode
		filesMock.On("Delete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()
		inodesMock.On("HardDelete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()

		// We remove the dir itself
		inodesMock.On("HardDelete", mock.Anything, inodes.ExampleAliceRoot.ID()).Return(nil).Once()

		svc := NewJob(inodesMock, filesMock, foldersMock, tools)

		err := svc.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("with a GetAllDeleted error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewJob(inodesMock, filesMock, foldersMock, tools)

		err := svc.Run(ctx)
		assert.EqualError(t, err, "failed to GetAllDeleted: some-error")
	})

	t.Run("with a Readdir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesMock.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewJob(inodesMock, filesMock, foldersMock, tools)

		err := svc.Run(ctx)
		assert.EqualError(t, err, "failed to delete inode \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\": failed to Readdir: some-error")
	})
}
