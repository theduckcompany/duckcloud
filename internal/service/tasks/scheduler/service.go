package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type TasksService struct {
	storage storage.Storage

	uuid  uuid.Service
	clock clock.Clock
}

func NewService(storage storage.Storage, tools tools.Tools) *TasksService {
	return &TasksService{storage, tools.UUID(), tools.Clock()}
}

func (t *TasksService) RegisterFileUploadTask(ctx context.Context, args *FileUploadArgs) error {
	err := args.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	return t.registerTask(ctx, 2, "file-upload", args)
}

func (t *TasksService) RegisterFSMove(ctx context.Context, args *FSMoveArgs) error {
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
