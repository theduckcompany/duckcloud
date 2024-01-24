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

type SpaceCreateTaskRunner struct {
	users  users.Service
	spaces spaces.Service
	fs     dfs.Service
}

func NewSpaceCreateTaskRunner(users users.Service, spaces spaces.Service, fs dfs.Service) *SpaceCreateTaskRunner {
	return &SpaceCreateTaskRunner{users, spaces, fs}
}

func (r *SpaceCreateTaskRunner) Name() string { return "space-create" }

func (r *SpaceCreateTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.SpaceCreateArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *SpaceCreateTaskRunner) RunArgs(ctx context.Context, args *scheduler.SpaceCreateArgs) error {
	user, err := r.users.GetByID(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to get the user by id (%q): %w", args.UserID, err)
	}

	space, err := r.spaces.Create(ctx, &spaces.CreateCmd{
		User:   user,
		Name:   args.Name,
		Owners: args.Owners,
	})
	if err != nil {
		return fmt.Errorf("failed to create the space: %w", err)
	}

	ctx = context.WithoutCancel(ctx)

	_, err = r.fs.CreateFS(ctx, user, space)
	if err != nil {
		return fmt.Errorf("failed to create the fs for space %q: %w", space.Name(), err)
	}

	return nil
}
