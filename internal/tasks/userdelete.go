package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
)

type UserDeleteTaskRunner struct {
	users         users.Service
	webSessions   websessions.Service
	davSessions   davsessions.Service
	oauthSessions oauthsessions.Service
	oauthConsents oauthconsents.Service
	spaces        spaces.Service
	fs            dfs.Service
}

func NewUserDeleteTaskRunner(
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	spaces spaces.Service,
	fs dfs.Service,
) *UserDeleteTaskRunner {
	return &UserDeleteTaskRunner{
		users,
		webSessions,
		davSessions,
		oauthSessions,
		oauthConsents,
		spaces,
		fs,
	}
}

func (r *UserDeleteTaskRunner) Name() string { return "user-delete" }

func (j *UserDeleteTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.UserDeleteArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return j.RunArgs(ctx, &args)
}

func (r *UserDeleteTaskRunner) RunArgs(ctx context.Context, args *scheduler.UserDeleteArgs) error {
	user, err := r.users.GetByID(ctx, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to find the user user %q: %w", args.UserID, err)
	}

	// First delete the accesses
	err = r.webSessions.DeleteAll(ctx, args.UserID)
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

	userSpaces, err := r.spaces.GetAllUserSpaces(ctx, args.UserID, nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllUserSpaces: %w", err)
	}

	for _, space := range userSpaces {
		_, err := r.spaces.RemoveOwner(ctx, &spaces.RemoveOwnerCmd{
			User:    user,
			Owner:   user,
			SpaceID: space.ID(),
		})
		if err != nil {
			return fmt.Errorf("failed to remove the user %q from the space %q: %w", user.ID(), space.ID(), err)
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
