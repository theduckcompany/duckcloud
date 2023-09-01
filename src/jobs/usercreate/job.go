package usercreate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

const (
	batchSize = 10
	jobName   = "usercreate"
)

type Job struct {
	users  users.Service
	inodes inodes.Service
	log    *slog.Logger
}

func NewJob(
	users users.Service,
	inodes inodes.Service,
	tools tools.Tools,
) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{
		users,
		inodes,
		logger,
	}
}

func (j *Job) Run(ctx context.Context) error {
	j.log.DebugContext(ctx, "start job")
	for {
		users, err := j.users.GetAllWithStatus(ctx, "initializing", &storage.PaginateCmd{Limit: batchSize})
		if err != nil {
			return fmt.Errorf("failed to GetAllWithStatus: %w", err)
		}

		for _, user := range users {
			err = j.bootstrapUser(ctx, &user)
			if err != nil {
				j.log.ErrorContext(ctx, "failed to bootstrap user", slog.String("error", err.Error()), slog.String("userID", string(user.ID())))
			}

			j.log.DebugContext(ctx, "user successfully deleted", slog.String("user", string(user.ID())))
		}

		if len(users) < batchSize {
			j.log.DebugContext(ctx, "end job")
			return nil
		}
	}
}

func (j *Job) bootstrapUser(ctx context.Context, user *users.User) error {
	// TODO: This action is not idempotent and could lead to orphan root dirs.
	rootDir, err := j.inodes.CreateRootDir(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to bootstrap the user inodes: %w", err)
	}

	_, err = j.users.SaveBootstrapInfos(ctx, user.ID(), rootDir)
	if err != nil {
		return fmt.Errorf("failed to SaveBootstrapInfos: %w", err)
	}

	return nil
}
