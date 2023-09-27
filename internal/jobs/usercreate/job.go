package usercreate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const (
	batchSize = 10
	jobName   = "usercreate"
)

type Job struct {
	users   users.Service
	folders folders.Service
	log     *slog.Logger
}

func NewJob(users users.Service, folders folders.Service, tools tools.Tools) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{users, folders, logger}
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
	folder, err := j.folders.CreatePersonalFolder(ctx, &folders.CreatePersonalFolderCmd{
		Name:  "My files",
		Owner: user.ID(),
	})
	if err != nil {
		return fmt.Errorf("failed to CreatePersonalFolder: %w", err)
	}

	_, err = j.users.SetDefaultFolder(ctx, *user, folder)
	if err != nil {
		return fmt.Errorf("failed to SetDefaultFolder: %w", err)
	}

	_, err = j.users.MarkInitAsFinished(ctx, user.ID())
	if err != nil {
		return fmt.Errorf("failed to MarkInitAsFinished: %w", err)
	}

	return nil
}
