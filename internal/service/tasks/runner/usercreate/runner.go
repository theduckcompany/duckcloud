package usercreate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
)

type TaskRunner struct {
	users   users.Service
	folders folders.Service
}

func NewTaskRunner(users users.Service, folders folders.Service) *TaskRunner {
	return &TaskRunner{users, folders}
}

func (r *TaskRunner) Name() string { return model.UserCreate }

func (r *TaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.UserCreateArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *TaskRunner) RunArgs(ctx context.Context, args *scheduler.UserCreateArgs) error {
	user, err := r.users.GetByID(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to retrieve the user %q: %w", args.UserID, err)
	}

	switch user.Status() {
	case users.Initializing:
		// This is expected, continue
	case users.Active, users.Deleting:
		// Already initialized or inside the deletion process, do nothing
		return nil
	default:
		return fmt.Errorf("unepected user status: %s", user.Status())
	}

	folder, err := r.folders.CreatePersonalFolder(ctx, &folders.CreatePersonalFolderCmd{
		Name:  "My files",
		Owner: args.UserID,
	})
	if err != nil {
		return fmt.Errorf("failed to CreatePersonalFolder: %w", err)
	}

	_, err = r.users.SetDefaultFolder(ctx, *user, folder)
	if err != nil {
		return fmt.Errorf("failed to SetDefaultFolder: %w", err)
	}

	_, err = r.users.MarkInitAsFinished(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to MarkInitAsFinished: %w", err)
	}

	return nil
}
