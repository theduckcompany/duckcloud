package fsgc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

const gcBatchSize = 10

type Job struct {
	inodes inodes.Service
	log    *slog.Logger
	cancel context.CancelFunc
	quit   chan struct{}
}

func NewJob(inodes inodes.Service, tools tools.Tools) *Job {
	return &Job{inodes, tools.Logger(), nil, make(chan struct{})}
}

func (j *Job) Start(pauseDuration time.Duration) {
	ticker := time.NewTicker(pauseDuration)
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel

	go func() {
		for {
			select {
			case <-ticker.C:
				err := j.run(ctx)
				if err != nil {
					j.log.Error("fs gc error", slog.String("error", err.Error()))
				}
			case <-j.quit:
				ticker.Stop()
				cancel()
			}
		}
	}()
}

func (j *Job) Stop() {
	close(j.quit)

	if j.cancel != nil {
		j.cancel()
	}
}

func (j *Job) run(ctx context.Context) error {
	for {
		toDelete, err := j.inodes.GetAllDeleted(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetAllDeleted: %w", err)
		}

		for _, inode := range toDelete {
			err = j.deleteINode(ctx, &inode)
			if err != nil {
				return fmt.Errorf("failed to delete inode %q: %w", inode.ID(), err)
			}

			j.log.DebugContext(ctx, "inode successfully removed", slog.String("inode", string(inode.ID())))
		}

		if len(toDelete) < gcBatchSize {
			return nil
		}
	}
}

func (j *Job) deleteINode(ctx context.Context, inode *inodes.INode) error {
	if inode.Mode().IsDir() {
		childs, err := j.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     inode.ID(),
			UserID:   inode.UserID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10})
		if err != nil {
			return fmt.Errorf("failed to Readdir: %w", err)
		}

		for _, child := range childs {
			err = j.deleteINode(ctx, &child)
			if err != nil {
				return fmt.Errorf("failed to deleteINode %q: %w", child.ID(), err)
			}
		}
	}

	return j.inodes.HardDelete(ctx, inode.ID())
}
