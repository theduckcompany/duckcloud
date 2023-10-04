package fsgc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const (
	gcBatchSize = 10
	jobName     = "fsgc"
)

type Job struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
	log     *slog.Logger
	cancel  context.CancelFunc
	clock   clock.Clock
	quit    chan struct{}
}

func NewJob(inodes inodes.Service, files files.Service, folders folders.Service, tools tools.Tools) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{inodes, files, folders, logger, nil, tools.Clock(), make(chan struct{})}
}

func (j *Job) Run(ctx context.Context) error {
	j.log.DebugContext(ctx, "start job")
	for {
		toDelete, err := j.inodes.GetAllDeleted(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetAllDeleted: %w", err)
		}

		for _, inode := range toDelete {
			deletionDate := inode.LastModifiedAt()

			err = j.deleteINode(ctx, &inode, deletionDate)
			if err != nil {
				return fmt.Errorf("failed to delete inode %q: %w", inode.ID(), err)
			}

			j.log.DebugContext(ctx, "inode successfully removed", slog.String("inode", string(inode.ID())))
		}

		if len(toDelete) < gcBatchSize {
			j.log.DebugContext(ctx, "end job")
			return nil
		}
	}
}

func (j *Job) deleteDirINode(ctx context.Context, inode *inodes.INode, deletionDate time.Time) error {
	for {
		childs, err := j.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     inode.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: gcBatchSize})
		if err != nil {
			return fmt.Errorf("failed to Readdir: %w", err)
		}

		for _, child := range childs {
			err = j.deleteINode(ctx, &child, j.clock.Now())
			if err != nil {
				return fmt.Errorf("failed to deleteINode %q: %w", child.ID(), err)
			}
		}

		if len(childs) < gcBatchSize {
			break
		}
	}

	err := j.inodes.HardDelete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	return nil
}

func (j *Job) deleteINode(ctx context.Context, inode *inodes.INode, deletionDate time.Time) error {
	if inode.Mode().IsDir() {
		return j.deleteDirINode(ctx, inode, deletionDate)
	}

	// For the file we have several steps:
	//
	// - Remove the inode
	// - Reduce all the parent folders size
	// - Remove the file
	err := j.inodes.HardDelete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	parentID := inode.Parent()
	for {
		if parentID == nil {
			break
		}

		parent, err := j.inodes.GetByID(ctx, *parentID)
		if err != nil {
			return fmt.Errorf("failed to GetByID the parent: %w", err)
		}

		if !parent.LastModifiedAt().Equal(deletionDate) {
			err = j.inodes.RegisterWrite(ctx, parent, -inode.Size(), deletionDate)
			if err != nil {
				return fmt.Errorf("failed to RegisterWrite: %w", err)
			}
		}

		parentID = parent.Parent()
	}

	err = j.files.Delete(ctx, inode)
	if err != nil {
		return fmt.Errorf("failed to remove the file %q: %w", inode.ID(), err)
	}

	return nil
}
