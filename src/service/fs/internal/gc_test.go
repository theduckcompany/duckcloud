package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestGC(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetDeletedINodes", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesSvc.On("Readdir", mock.Anything, &inodes.PathCmd{
			UserID:   inodes.ExampleAliceRoot.UserID(),
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		// We remove the content
		inodesSvc.On("HardDelete", mock.Anything, inodes.ExampleAliceFile.ID()).Return(nil).Once()
		// We remove the dir itself
		inodesSvc.On("HardDelete", mock.Anything, inodes.ExampleAliceRoot.ID()).Return(nil).Once()

		svc := NewGCService(inodesSvc, tools)

		err := svc.run(ctx)
		assert.NoError(t, err)
	})

	t.Run("with a GetDeletedINodes error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetDeletedINodes", mock.Anything, 10).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewGCService(inodesSvc, tools)

		err := svc.run(ctx)
		assert.EqualError(t, err, "failed to GetDeletedINodes: some-error")
	})

	t.Run("with a Readdir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)

		// First loop to fetch the deleted inodes
		inodesSvc.On("GetDeletedINodes", mock.Anything, 10).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// This is a dir we will delete all its content
		inodesSvc.On("Readdir", mock.Anything, &inodes.PathCmd{
			UserID:   inodes.ExampleAliceRoot.UserID(),
			Root:     inodes.ExampleAliceRoot.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10}).Return(nil, fmt.Errorf("some-error")).Once()

		svc := NewGCService(inodesSvc, tools)

		err := svc.run(ctx)
		assert.EqualError(t, err, "failed to delete inode \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\": failed to Readdir: some-error")
	})

	t.Run("Start an async job and stop it with Stop", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)

		svc := NewGCService(inodesSvc, tools)

		// Start the async job. The first call is done 1s after the call to Start
		svc.Start(time.Second)

		// Stop will interrupt the job before the second.
		svc.Stop()
	})

	t.Run("Stop interrupt the running job", func(t *testing.T) {
		tools := tools.NewMock(t)
		inodesSvc := inodes.NewMockService(t)

		svc := NewGCService(inodesSvc, tools)

		svc.Start(time.Millisecond)

		// First loop to fetch the deleted inodes. Make it take more than a 1s.
		inodesSvc.On("GetDeletedINodes", mock.Anything, 10).WaitUntil(time.After(time.Minute)).Return([]inodes.INode{inodes.ExampleAliceRoot}, nil).Once()

		// Wait some time in order to be just to have the job running and waiting for the end of "GetDeletedINodes".
		time.Sleep(20 * time.Millisecond)

		// Stop will interrupt the job before the second.
		start := time.Now()
		svc.Stop()
		assert.WithinDuration(t, time.Now(), start, 10*time.Millisecond)
	})
}
