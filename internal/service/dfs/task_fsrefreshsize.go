package dfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSRefreshSizeTaskRunner struct {
	inodes inodes.Service
	files  files.Service
}

func NewFSRefreshSizeTaskRunner(inodes inodes.Service, files files.Service) *FSRefreshSizeTaskRunner {
	return &FSRefreshSizeTaskRunner{inodes, files}
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
	inodeID := &args.INode

	for {
		if inodeID == nil {
			break
		}

		inode, err := r.inodes.GetByID(ctx, *inodeID)
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}

		var newSize uint64

		switch inode.IsDir() {
		case true:
			newSize, err = r.inodes.GetSumChildsSize(ctx, inode.ID())
			if err != nil {
				return fmt.Errorf("failed to get the total size for inode %q: %w", *inodeID, err)
			}

		case false:
			fileMeta, err := r.files.GetMetadata(ctx, *inode.FileID())
			if err != nil {
				return fmt.Errorf("failed to get the FileID for inode %q: %w", *inodeID, err)
			}
			newSize = fileMeta.Size()
		}

		err = r.inodes.RegisterModification(ctx, inode, newSize, args.ModifiedAt)
		if err != nil {
			return fmt.Errorf("failed to register the size modification: %w", err)
		}

		inodeID = inode.Parent()
	}

	return nil
}
