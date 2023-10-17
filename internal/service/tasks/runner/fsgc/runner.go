package fsgc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const gcBatchSize = 10

type TaskRunner struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
	cancel  context.CancelFunc
	clock   clock.Clock
	quit    chan struct{}
}

func NewTaskRunner(inodes inodes.Service, files files.Service, folders folders.Service, tools tools.Tools) *TaskRunner {
	return &TaskRunner{inodes, files, folders, nil, tools.Clock(), make(chan struct{})}
}

func (r *TaskRunner) Name() string { return model.FSGC }

func (r *TaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	return r.RunArgs(ctx, &scheduler.FSGCArgs{})
}

func (r *TaskRunner) RunArgs(ctx context.Context, args *scheduler.FSGCArgs) error {
	for {
		toDelete, err := r.inodes.GetAllDeleted(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetAllDeleted: %w", err)
		}

		for _, inode := range toDelete {
			deletionDate := inode.LastModifiedAt()

			err = r.deleteINode(ctx, &inode, deletionDate)
			if err != nil {
				return fmt.Errorf("failed to delete inode %q: %w", inode.ID(), err)
			}
		}

		if len(toDelete) < gcBatchSize {
			return nil
		}
	}
}

func (r *TaskRunner) deleteDirINode(ctx context.Context, inode *inodes.INode, deletionDate time.Time) error {
	for {
		childs, err := r.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     inode.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: gcBatchSize})
		if err != nil {
			return fmt.Errorf("failed to Readdir: %w", err)
		}

		for _, child := range childs {
			err = r.deleteINode(ctx, &child, r.clock.Now())
			if err != nil {
				return fmt.Errorf("failed to deleteINode %q: %w", child.ID(), err)
			}
		}

		if len(childs) < gcBatchSize {
			break
		}
	}

	err := r.inodes.HardDelete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	return nil
}

func (j *TaskRunner) deleteINode(ctx context.Context, inode *inodes.INode, deletionDate time.Time) error {
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

	err = j.files.Delete(ctx, *inode.FileID())
	if err != nil {
		return fmt.Errorf("failed to remove the file %q: %w", inode.ID(), err)
	}

	return nil
}
