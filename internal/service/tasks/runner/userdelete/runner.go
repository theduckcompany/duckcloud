package userdelete

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
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
	fs            dfs.Service
}

func NewTaskRunner(
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	folders folders.Service,
	fs dfs.Service,
) *TaskRunner {
	return &TaskRunner{
		users,
		webSessions,
		davSessions,
		oauthSessions,
		oauthConsents,
		folders,
		fs,
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

func (j *TaskRunner) RunArgs(ctx context.Context, args *scheduler.UserDeleteArgs) error {
	// First delete the accesses
	err := j.webSessions.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all web sessions: %w", err)
	}

	err = j.davSessions.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all dav sessions: %w", err)
	}

	err = j.oauthSessions.DeleteAllForUser(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all oauth sessions: %w", err)
	}

	folders, err := j.folders.GetAllUserFolders(ctx, args.UserID, nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllUserFolders: %w", err)
	}

	for _, folder := range folders {
		if folder.IsPublic() {
			continue
		}

		// Then delete the data
		ffs := j.fs.GetFolderFS(&folder)
		err = ffs.RemoveAll(ctx, "/")
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return fmt.Errorf("failed to delete the user root fs: %w", err)
		}

		err = j.folders.Delete(ctx, folder.ID())
		if err != nil {
			return fmt.Errorf("failed to delete the folder %q: %w", folder.ID(), err)
		}
	}

	err = j.oauthConsents.DeleteAll(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete all oauth consents: %w", err)
	}

	err = j.users.HardDelete(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to hard delete the user: %w", err)
	}

	return nil
}
