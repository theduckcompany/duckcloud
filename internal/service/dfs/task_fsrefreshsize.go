package dfs

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"

	inodes "github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSRefreshSizeTaskRunner struct {
	inodes inodes.Service
}

func NewFSRefreshSizeTaskRunner(inodes inodes.Service) *FSRefreshSizeTaskRunner {
	return &FSRefreshSizeTaskRunner{inodes}
}

func (r *FSRefreshSizeTaskRunner) Name() string { return "fs-refresh-size" }

func (r *FSRefreshSizeTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.FSRefreshSizeArg
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *FSRefreshSizeTaskRunner) RunArgs(ctx context.Context, args *scheduler.FSRefreshSizeArg) error {
	parentID := &args.INode

	for {
		if parentID == nil {
			break
		}

		parent, err := r.inodes.GetByID(ctx, *parentID)
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to GetByID the parent: %w", err)
		}

		parentSize, err := r.inodes.GetSumChildsSize(ctx, parent.ID())
		if err != nil {
			return fmt.Errorf("failed to get the total size for inode %q: %w", *parentID, err)
		}

		err = r.inodes.RegisterModification(ctx, parent, parentSize, args.ModifiedAt)
		if err != nil {
			return fmt.Errorf("failed to register the size modification: %w", err)
		}

		parentID = parent.Parent()
	}

	return nil
}
