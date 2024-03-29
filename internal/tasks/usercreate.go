package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
)

type UserCreateTaskRunner struct {
	users  users.Service
	spaces spaces.Service
	fs     dfs.Service
}

func NewUserCreateTaskRunner(users users.Service, spaces spaces.Service, fs dfs.Service) *UserCreateTaskRunner {
	return &UserCreateTaskRunner{users, spaces, fs}
}

func (r *UserCreateTaskRunner) Name() string { return "user-create" }

func (r *UserCreateTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.UserCreateArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *UserCreateTaskRunner) RunArgs(ctx context.Context, args *scheduler.UserCreateArgs) error {
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

	_, err = r.users.MarkInitAsFinished(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to MarkInitAsFinished: %w", err)
	}

	return nil
}
