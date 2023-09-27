package userdelete

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const (
	gcBatchSize = 10
	jobName     = "userdelete"
)

type Job struct {
	users         users.Service
	webSessions   websessions.Service
	davSessions   davsessions.Service
	oauthSessions oauthsessions.Service
	oauthConsents oauthconsents.Service
	folders       folders.Service
	fs            fs.Service
	log           *slog.Logger
}

func NewJob(
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	folders folders.Service,
	fs fs.Service,
	tools tools.Tools,
) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{
		users,
		webSessions,
		davSessions,
		oauthSessions,
		oauthConsents,
		folders,
		fs,
		logger,
	}
}

func (j *Job) Run(ctx context.Context) error {
	j.log.DebugContext(ctx, "start job")
	for {
		users, err := j.users.GetAllWithStatus(ctx, "deleting", &storage.PaginateCmd{Limit: gcBatchSize})
		if err != nil {
			return fmt.Errorf("failed to GetAllWithStatus: %w", err)
		}

		for _, user := range users {
			err = j.deleteUser(ctx, &user)
			if err != nil {
				return fmt.Errorf("failed to delete user %q: %w", user.ID(), err)
			}

			j.log.DebugContext(ctx, "user successfully deleted", slog.String("user", string(user.ID())))
		}

		if len(users) < gcBatchSize {
			j.log.DebugContext(ctx, "end job")
			return nil
		}
	}
}

func (j *Job) deleteUser(ctx context.Context, user *users.User) error {
	// First delete the accesses
	err := j.webSessions.DeleteAll(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to delete all web sessions: %w", err)
	}

	err = j.davSessions.DeleteAll(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to delete all dav sessions: %w", err)
	}

	err = j.oauthSessions.DeleteAllForUser(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to delete all oauth sessions: %w", err)
	}

	folders, err := j.folders.GetAllUserFolders(ctx, user.ID(), nil)
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

	err = j.oauthConsents.DeleteAll(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to delete all oauth consents: %w", err)
	}

	err = j.users.HardDelete(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to hard delete the user: %w", err)
	}

	return nil
}
