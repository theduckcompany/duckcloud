package dfs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const gcBatchSize = 10

type FSGGCTaskRunner struct {
	storage Storage
	files   files.Service
	spaces  spaces.Service
	cancel  context.CancelFunc
	clock   clock.Clock
	quit    chan struct{}
}

func NewFSGGCTaskRunner(
	storage Storage,
	files files.Service,
	spaces spaces.Service,
	tools tools.Tools,
) *FSGGCTaskRunner {
	return &FSGGCTaskRunner{storage, files, spaces, nil, tools.Clock(), make(chan struct{})}
}

func (r *FSGGCTaskRunner) Name() string { return "fs-gc" }

func (r *FSGGCTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	return r.RunArgs(ctx, &scheduler.FSGCArgs{})
}

func (r *FSGGCTaskRunner) RunArgs(ctx context.Context, args *scheduler.FSGCArgs) error {
	for {
		toDelete, err := r.storage.GetAllDeleted(ctx, gcBatchSize)
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

func (r *FSGGCTaskRunner) deleteDirINode(ctx context.Context, inode *INode, deletionDate time.Time) error {
	for {
		childs, err := r.storage.GetAllChildrens(ctx, inode.ID(), &storage.PaginateCmd{Limit: gcBatchSize})
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

	err := r.storage.HardDelete(ctx, inode.id)
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	return nil
}

func (j *FSGGCTaskRunner) deleteINode(ctx context.Context, inode *INode, deletionDate time.Time) error {
	// XXX:MULTI-WRITE
	//
	// This file have severa consecutive writes but they are all idempotent and the
	// task is retried in case of error.
	if inode.IsDir() {
		return j.deleteDirINode(ctx, inode, deletionDate)
	}

	err := j.storage.HardDelete(ctx, inode.id)
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	inodes, err := j.storage.GetAllInodesWithFileID(ctx, *inode.FileID())
	if err != nil {
		return fmt.Errorf("failed to GetAllINodesWithFileID: %w", err)
	}

	if len(inodes) == 0 {
		// No more inodes target this file so it can be removed
		err = j.files.Delete(ctx, *inode.FileID())
		if err != nil {
			return fmt.Errorf("failed to remove the file %q: %w", inode.ID(), err)
		}
	}

	return nil
}
