package userdelete

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
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
	inodes        inodes.Service
	log           *slog.Logger
}

func NewJob(
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	inodes inodes.Service,
	tools tools.Tools,
) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{
		users,
		webSessions,
		davSessions,
		oauthSessions,
		oauthConsents,
		inodes,
		logger,
	}
}

func (j *Job) Run(ctx context.Context) error {
	j.log.DebugContext(ctx, "start job")
	for {
		users, err := j.users.GetAllDeleted(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetAllDeleted: %w", err)
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

	// Then delete the data
	err = j.inodes.RemoveAll(ctx, &inodes.PathCmd{
		Root:     user.RootFS(),
		UserID:   user.ID(),
		FullName: "/",
	})
	if err != nil {
		return fmt.Errorf("failed to delete the user root fs: %w", err)
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
