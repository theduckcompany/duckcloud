package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSchdulerService(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2020, time.February, 12, 11, 10, 0, 0, time.UTC)

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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
			Args:         json.RawMessage(`{"folder-id":"a379fef3-ebc3-4069-b1ef-8c67948b3cff","file-id":"0d76c071-2e8b-4873-92e9-d8be871ef636","path":"/foo/bar.txt","uploaded-at":"2020-02-12T11:10:00Z"}`),
		}).Return(nil).Once()

		err := svc.RegisterFileUploadTask(ctx, &FileUploadArgs{
			FolderID:   uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			FileID:     uuid.UUID("0d76c071-2e8b-4873-92e9-d8be871ef636"),
			Path:       "/foo/bar.txt",
			UploadedAt: now,
		})
		assert.NoError(t, err)
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
			Args:         json.RawMessage(`{"folder":"a379fef3-ebc3-4069-b1ef-8c67948b3cff","source-inode":"0d76c071-2e8b-4873-92e9-d8be871ef636","target-path":"/foo/bar.txt","moved-at":"2020-02-12T11:10:00Z"}`),
		}).Return(nil).Once()

		err := svc.RegisterFSMoveTask(ctx, &FSMoveArgs{
			FolderID:    uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			SourceInode: uuid.UUID("0d76c071-2e8b-4873-92e9-d8be871ef636"),
			TargetPath:  "/foo/bar.txt",
			MovedAt:     now,
		})
		assert.NoError(t, err)
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
