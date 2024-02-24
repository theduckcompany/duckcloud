package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSchdulerService(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2020, time.February, 12, 11, 10, 0, 0, time.UTC)

	t.Run("RegisterSpaceCreateTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     1,
			Status:       model.Queuing,
			Name:         "space-create",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"user-id":"ad7dfbfd-fed3-4ee0-8d61-5321986aa6b0","name":"First Space","owners":["fac2e836-eb4b-4eba-8184-25c332180326"]}`),
		}).Return(nil).Once()

		err := svc.RegisterSpaceCreateTask(ctx, &SpaceCreateArgs{
			UserID: uuid.UUID("ad7dfbfd-fed3-4ee0-8d61-5321986aa6b0"),
			Name:   "First Space",
			Owners: []uuid.UUID{uuid.UUID("fac2e836-eb4b-4eba-8184-25c332180326")},
		})
		require.NoError(t, err)
	})

	t.Run("RegisterSpaceCreateTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		err := svc.RegisterSpaceCreateTask(ctx, &SpaceCreateArgs{
			UserID: uuid.UUID("invalid"),
			Name:   "First Space",
			Owners: []uuid.UUID{uuid.UUID("fac2e836-eb4b-4eba-8184-25c332180326")},
		})
		require.ErrorIs(t, err, errs.ErrValidation)
	})

	t.Run("RegisterFSGCTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     4,
			Status:       model.Queuing,
			Name:         "fs-gc",
			RegisteredAt: now,
			Args:         json.RawMessage(`{}`),
		}).Return(nil).Once()

		err := svc.RegisterFSGCTask(ctx)
		require.NoError(t, err)
	})

	t.Run("RegisterUserDeleteTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     1,
			Status:       model.Queuing,
			Name:         "user-delete",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"user-id":"a379fef3-ebc3-4069-b1ef-8c67948b3cff"}`),
		}).Return(nil).Once()

		err := svc.RegisterUserDeleteTask(ctx, &UserDeleteArgs{
			UserID: uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
		})
		require.NoError(t, err)
	})

	t.Run("RegisterUserCreateTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     1,
			Status:       model.Queuing,
			Name:         "user-create",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"user-id":"a379fef3-ebc3-4069-b1ef-8c67948b3cff"}`),
		}).Return(nil).Once()

		err := svc.RegisterUserCreateTask(ctx, &UserCreateArgs{
			UserID: uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
		})
		require.NoError(t, err)
	})

	t.Run("RegisterFileUploadTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     2,
			Status:       model.Queuing,
			Name:         "file-upload",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"uploaded-at":"2020-02-12T11:10:00Z","space-id":"a379fef3-ebc3-4069-b1ef-8c67948b3cff","file-id":"0d76c071-2e8b-4873-92e9-d8be871ef636","inode-id":"c87ebbda-435b-43b7-bab6-e93ca8f3831a"}`),
		}).Return(nil).Once()

		err := svc.RegisterFileUploadTask(ctx, &FileUploadArgs{
			SpaceID:    uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			FileID:     uuid.UUID("0d76c071-2e8b-4873-92e9-d8be871ef636"),
			INodeID:    uuid.UUID("c87ebbda-435b-43b7-bab6-e93ca8f3831a"),
			UploadedAt: now,
		})
		require.NoError(t, err)
	})

	t.Run("RegisterFSRefreshSizeTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     2,
			Status:       model.Queuing,
			Name:         "fs-refresh-size",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"modified_at":"2020-02-12T11:10:00Z","inode":"a379fef3-ebc3-4069-b1ef-8c67948b3cff"}`),
		}).Return(nil).Once()

		err := svc.RegisterFSRefreshSizeTask(ctx, &FSRefreshSizeArg{
			INode:      uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			ModifiedAt: now,
		})
		require.NoError(t, err)
	})

	t.Run("RegisterFSMoveTask", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid")).Once()
		tools.ClockMock.On("Now").Return(now).Once()

		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     2,
			Status:       model.Queuing,
			Name:         "fs-move",
			RegisteredAt: now,
			Args:         json.RawMessage(`{"space":"a379fef3-ebc3-4069-b1ef-8c67948b3cff","source-inode":"0d76c071-2e8b-4873-92e9-d8be871ef636","target-path":"/foo/bar.txt","moved-at":"2020-02-12T11:10:00Z","moved-by":"74926c6a-1802-45cd-bcb2-2dc0729fa986"}`),
		}).Return(nil).Once()

		err := svc.RegisterFSMoveTask(ctx, &FSMoveArgs{
			SpaceID:     uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			SourceInode: uuid.UUID("0d76c071-2e8b-4873-92e9-d8be871ef636"),
			TargetPath:  "/foo/bar.txt",
			MovedAt:     now,
			MovedBy:     uuid.UUID("74926c6a-1802-45cd-bcb2-2dc0729fa986"),
		})
		require.NoError(t, err)
	})

	t.Run("Run success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		storageMock.On("GetLastRegisteredTask", mock.Anything, "fs-gc").Return(&model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     4,
			Status:       model.Queuing,
			Name:         "fs-gc",
			RegisteredAt: now.Add(-time.Hour),
			Args:         json.RawMessage(`{}`),
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		// There is 1H since the last "fs-gc" task so it will create a new one

		tools.UUIDMock.On("New").Return(uuid.UUID("some-new-uuid")).Once()
		tools.ClockMock.On("Now").Return(now.Add(time.Second)).Once()
		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-new-uuid"),
			Priority:     4,
			Status:       model.Queuing,
			Name:         "fs-gc",
			RegisteredAt: now.Add(time.Second),
			Args:         json.RawMessage(`{}`),
		}).Return(nil).Once()

		err := svc.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run with a task done a fews seconds ago", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		storageMock.On("GetLastRegisteredTask", mock.Anything, "fs-gc").Return(&model.Task{
			ID:           uuid.UUID("some-uuid"),
			Priority:     4,
			Status:       model.Queuing,
			Name:         "fs-gc",
			RegisteredAt: now.Add(-3 * time.Second),
			Args:         json.RawMessage(`{}`),
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		// There is 2 seconds since the last "fs-gc" task so there is no need
		// to push a new task.

		err := svc.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run with now tasks", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := storage.NewMockStorage(t)
		svc := NewService(storageMock, tools)

		storageMock.On("GetLastRegisteredTask", mock.Anything, "fs-gc").Return(nil, storage.ErrNotFound).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-new-uuid")).Once()
		tools.ClockMock.On("Now").Return(now.Add(time.Second)).Once()
		storageMock.On("Save", mock.Anything, &model.Task{
			ID:           uuid.UUID("some-new-uuid"),
			Priority:     4,
			Status:       model.Queuing,
			Name:         "fs-gc",
			RegisteredAt: now.Add(time.Second),
			Args:         json.RawMessage(`{}`),
		}).Return(nil).Once()

		err := svc.Run(ctx)
		require.NoError(t, err)
	})
}
