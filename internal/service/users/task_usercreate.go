package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type UserCreateTaskRunner struct {
	users   Service
	folders folders.Service
	fs      dfs.Service
}

func NewUserCreateTaskRunner(users Service, folders folders.Service, fs dfs.Service) *UserCreateTaskRunner {
	return &UserCreateTaskRunner{users, folders, fs}
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
	case Initializing:
		// This is expected, continue
	case Active, Deleting:
		// Already initialized or inside the deletion process, do nothing
		return nil
	default:
		return fmt.Errorf("unepected user status: %s", user.Status())
	}

	existingFolders, err := r.folders.GetAllUserFolders(ctx, user.ID(), nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllUserFolders: %w", err)
	}

	var firstFolder *folders.Folder
	switch len(existingFolders) {
	case 0:
		firstFolder = nil
	case 1:
		firstFolder = &existingFolders[0]
	default:
		return fmt.Errorf("the new user already have several folder: %+v", existingFolders)
	}

	if firstFolder == nil {
		firstFolder, err = r.fs.CreateFS(ctx, []uuid.UUID{user.ID()})
		if err != nil {
			return fmt.Errorf("failed to CreateFS: %w", err)
		}
	}

	_, err = r.users.SetDefaultFolder(ctx, *user, firstFolder)
	if err != nil {
		return fmt.Errorf("failed to SetDefaultFolder: %w", err)
	}

	_, err = r.users.MarkInitAsFinished(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to MarkInitAsFinished: %w", err)
	}

	return nil
}
