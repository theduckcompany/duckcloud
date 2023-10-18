package usercreate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type TaskRunner struct {
	users   users.Service
	folders folders.Service
	inodes  inodes.Service
}

func NewTaskRunner(users users.Service, folders folders.Service, inodes inodes.Service) *TaskRunner {
	return &TaskRunner{users, folders, inodes}
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
		rootFS, err := r.inodes.CreateRootDir(ctx)
		if err != nil {
			return fmt.Errorf("failed to CreateRootDir: %w", err)
		}

		// XXX:MULTI-WRITE
		firstFolder, err = r.folders.Create(ctx, &folders.CreateCmd{
			Name:   "My files",
			Owners: []uuid.UUID{args.UserID},
			RootFS: rootFS.ID(),
		})
		if err != nil {
			return fmt.Errorf("failed to CreatePersonalFolder: %w", err)
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
