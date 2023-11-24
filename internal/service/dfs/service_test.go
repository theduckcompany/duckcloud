package dfs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDFSService(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateSpaceFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, schedulerMock, toolsMock)

		inodesMock.On("CreateSpaceRootDir", mock.Anything, uuid.UUID("some-space-id")).Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := svc.CreateSpaceFS(ctx, uuid.UUID("some-space-id"))
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceRoot, res)
	})

	t.Run("CreateSpaceFS with a create RootDirError", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, schedulerMock, toolsMock)

		inodesMock.On("CreateSpaceRootDir", mock.Anything, uuid.UUID("some-space-id")).Return(nil, errs.Internal(fmt.Errorf("some-error"))).Once()

		res, err := svc.CreateSpaceFS(ctx, uuid.UUID("some-space-id"))
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveSpaceFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, schedulerMock, toolsMock)

		inodesMock.On("GetSpaceRoot", mock.Anything, uuid.UUID("some-space-id")).Return(&ExampleAliceRoot, nil).Once()
		inodesMock.On("Remove", mock.Anything, &ExampleAliceRoot).Return(nil).Once()

		err := svc.RemoveSpaceFS(ctx, uuid.UUID("some-space-id"))
		assert.NoError(t, err)
	})

	t.Run("RemoveSpaceFS with an Remove error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		inodesMock := inodes.NewMockService(t)
		filesMock := files.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		svc := NewFSService(inodesMock, filesMock, schedulerMock, toolsMock)

		inodesMock.On("GetSpaceRoot", mock.Anything, uuid.UUID("some-space-id")).Return(&ExampleAliceRoot, nil).Once()
		inodesMock.On("Remove", mock.Anything, &ExampleAliceRoot).
			Return(errs.Internal(errors.New("some-error"))).Once()

		err := svc.RemoveSpaceFS(ctx, uuid.UUID("some-space-id"))
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
