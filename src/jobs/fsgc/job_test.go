package fsgc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestFSGC(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesSvc.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		// We remove the content
		filesMock.On("Delete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()
		inodesSvc.On("HardDelete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()
		// We remove the dir itself
		inodesSvc.On("HardDelete", mock.Anything, inodes.ExampleAliceRoot.ID()).Return(nil).Once()

		svc := NewJob(inodesSvc, filesMock, tools)

		err := svc.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("with a GetAllDeleted error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetAllDeleted", mock.Anything, 10).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewJob(inodesSvc, filesMock, tools)

		err := svc.Run(ctx)
		assert.EqualError(t, err, "failed to GetAllDeleted: some-error")
	})

	t.Run("with a Readdir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetAllDeleted", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesSvc.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewJob(inodesSvc, filesMock, tools)

		err := svc.Run(ctx)
		assert.EqualError(t, err, "failed to delete inode \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\": failed to Readdir: some-error")
	})
}
