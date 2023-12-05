package dfs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDFSService(t *testing.T) {
	ctx := context.Background()

	// Copy the id to avoid a dependency cycle
	AliceUserID := uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	t.Run("CreateFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   DefaultSpaceName,
			Owners: []uuid.UUID{AliceUserID},
		}).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceRoot.ID())
		storageMock.On("Save", mock.Anything, &INode{
			id:             ExampleAliceRoot.id,
			parent:         nil,
			name:           "",
			spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
			createdAt:      now,
			createdBy:      users.ExampleAlice.ID(),
			lastModifiedAt: now,
			fileID:         nil,
		}).Return(nil)

		res, err := svc.CreateFS(ctx, &users.ExampleAlice, []uuid.UUID{AliceUserID})
		assert.NoError(t, err)
		assert.Equal(t, &spaces.ExampleAlicePersonalSpace, res)
	})

	t.Run("CreateFS with a create storage error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   DefaultSpaceName,
			Owners: []uuid.UUID{AliceUserID},
		}).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		toolsMock.ClockMock.On("Now").Return(now).Once()
		toolsMock.UUIDMock.On("New").Return(ExampleAliceRoot.ID())
		storageMock.On("Save", mock.Anything, &INode{
			id:             ExampleAliceRoot.id,
			parent:         nil,
			name:           "",
			spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
			createdAt:      now,
			createdBy:      users.ExampleAlice.ID(),
			lastModifiedAt: now,
			fileID:         nil,
		}).Return(fmt.Errorf("some-error"))

		res, err := svc.CreateFS(ctx, &users.ExampleAlice, []uuid.UUID{AliceUserID})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("CreateFS with a space create error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   DefaultSpaceName,
			Owners: []uuid.UUID{AliceUserID},
		}).Return(nil, errs.Internal(errors.New("some-error"))).Once()

		res, err := svc.CreateFS(ctx, &users.ExampleAlice, []uuid.UUID{AliceUserID})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS success", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		// Delete the file system
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&ExampleAliceRoot, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()

		// Delete the space
		spacesMock.On("Delete", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).Return(nil).Once()

		err := svc.RemoveFS(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.NoError(t, err)
	})

	t.Run("RemoveFS with a GetSpaceRoot error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		// Delete the file system
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(nil, fmt.Errorf("some-error")).Once()

		err := svc.RemoveFS(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS with a Patch error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		// Delete the file system
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&ExampleAliceRoot, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(fmt.Errorf("some-error")).Once()

		err := svc.RemoveFS(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})

	t.Run("RemoveFS with a delete space error", func(t *testing.T) {
		toolsMock := tools.NewMock(t)
		filesMock := files.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		storageMock := NewMockStorage(t)
		svc := NewFSService(storageMock, filesMock, spacesMock, schedulerMock, toolsMock)

		// Delete the file system
		storageMock.On("GetSpaceRoot", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(&ExampleAliceRoot, nil).Once()
		toolsMock.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceRoot.ID(), map[string]any{
			"deleted_at":       now,
			"last_modified_at": now,
		}).Return(nil).Once()

		// Delete the space
		spacesMock.On("Delete", mock.Anything, spaces.ExampleAlicePersonalSpace.ID()).
			Return(errs.Internal(fmt.Errorf("some-error"))).Once()

		err := svc.RemoveFS(ctx, &spaces.ExampleAlicePersonalSpace)
		assert.ErrorIs(t, err, errs.ErrInternal)
		assert.ErrorContains(t, err, "some-error")
	})
}
