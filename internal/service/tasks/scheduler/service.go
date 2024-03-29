package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/taskstorage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type TasksService struct {
	storage taskstorage.Storage

	uuid  uuid.Service
	clock clock.Clock
}

func NewService(storage taskstorage.Storage, tools tools.Tools) *TasksService {
	return &TasksService{storage, tools.UUID(), tools.Clock()}
}

func (t *TasksService) Run(ctx context.Context) error {
	err := t.ensureTaskEvery(ctx, "fs-gc", 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to schedule fs-gc task: %w", err)
	}

	return nil
}

func (t *TasksService) ensureTaskEvery(ctx context.Context, name string, every time.Duration) error {
	lastTask, err := t.storage.GetLastRegisteredTask(ctx, name)
	if err != nil && !errors.Is(err, taskstorage.ErrNotFound) {
		return fmt.Errorf("failed to GetLastRegisteredTask: %w", err)
	}

	if lastTask != nil && lastTask.RegisteredAt.Add(every).After(t.clock.Now()) {
		return nil
	}

	switch name {
	case "fs-gc":
		return t.RegisterFSGCTask(ctx)
	default:
		return fmt.Errorf("unhandled task name")
	}
}

func (t *TasksService) RegisterFSGCTask(ctx context.Context) error {
	return t.registerTask(ctx, 4, "fs-gc", struct{}{})
}

func (t *TasksService) RegisterFSRefreshSizeTask(ctx context.Context, args *FSRefreshSizeArg) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 2, "fs-refresh-size", args)
}

func (t *TasksService) RegisterFSRemoveDuplicateFile(ctx context.Context, args *FSRemoveDuplicateFileArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 2, "fs-remove-duplicate-file", args)
}

func (t *TasksService) RegisterFileUploadTask(ctx context.Context, args *FileUploadArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 2, "file-upload", args)
}

func (t *TasksService) RegisterFSMoveTask(ctx context.Context, args *FSMoveArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 2, "fs-move", args)
}

func (t *TasksService) RegisterUserCreateTask(ctx context.Context, args *UserCreateArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 1, "user-create", args)
}

func (t *TasksService) RegisterUserDeleteTask(ctx context.Context, args *UserDeleteArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 1, "user-delete", args)
}

func (t *TasksService) RegisterSpaceCreateTask(ctx context.Context, args *SpaceCreateArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 1, "space-create", args)
}

func (t *TasksService) registerTask(ctx context.Context, priority int, name string, args any) error {
	rawArgs, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal the args: %w", err)
	}

	err = t.storage.Save(ctx, &model.Task{
		ID:           t.uuid.New(),
		Priority:     priority,
		Status:       model.Queuing,
		Name:         name,
		RegisteredAt: t.clock.Now(),
		Args:         rawArgs,
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to save the %q job : %w", name, err))
	}

	return nil
}
