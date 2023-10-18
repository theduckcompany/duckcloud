package userdelete

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type TaskRunner struct {
	users         users.Service
	webSessions   websessions.Service
	davSessions   davsessions.Service
	oauthSessions oauthsessions.Service
	oauthConsents oauthconsents.Service
	folders       folders.Service
	inodes        inodes.Service
}

func NewTaskRunner(
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	folders folders.Service,
	inodes inodes.Service,
) *TaskRunner {
	return &TaskRunner{
		users,
		webSessions,
		davSessions,
		oauthSessions,
		oauthConsents,
		folders,
		inodes,
	}
}

func (r *TaskRunner) Name() string { return model.UserDelete }

func (j *TaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.UserDeleteArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return j.RunArgs(ctx, &args)
}

func (r *TaskRunner) RunArgs(ctx context.Context, args *scheduler.UserDeleteArgs) error {
	// First delete the accesses
	err := r.webSessions.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all web sessions: %w", err)
	}

	err = r.davSessions.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all dav sessions: %w", err)
	}

	err = r.oauthSessions.DeleteAllForUser(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all oauth sessions: %w", err)
	}

	folders, err := r.folders.GetAllUserFolders(ctx, args.UserID, nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllUserFolders: %w", err)
	}

	for _, folder := range folders {
		if folder.IsPublic() {
			continue
		}

		rootFS, err := r.inodes.GetByID(ctx, folder.RootFS())
		if errors.Is(err, errs.ErrNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to Get the rootFS for %q: %w", folder.Name(), err)
		}

		// TODO: Create a folderdelete task
		err = r.inodes.Remove(ctx, rootFS)
		if err != nil {
			return fmt.Errorf("failed to remove the rootFS for %q: %w", folder.Name(), err)
		}

		err = r.folders.Delete(ctx, folder.ID())
		if err != nil {
			return fmt.Errorf("failed to delete the folder %q: %w", folder.ID(), err)
		}
	}

	err = r.oauthConsents.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all oauth consents: %w", err)
	}

	err = r.users.HardDelete(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to hard delete the user: %w", err)
	}

	return nil
}
