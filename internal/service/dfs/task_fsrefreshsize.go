package dfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSRefreshSizeTaskRunner struct {
	storage Storage
	files   files.Service
}

func NewFSRefreshSizeTaskRunner(storage Storage, files files.Service) *FSRefreshSizeTaskRunner {
	return &FSRefreshSizeTaskRunner{storage, files}
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
	var newSize uint64
	inodeID := &args.INode

	for inodeID != nil {
		inode, err := r.storage.GetByID(ctx, *inodeID)
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}

		switch inode.IsDir() {
		case true:
			newSize, err = r.storage.GetSumChildsSize(ctx, inode.ID())
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

		err = r.storage.Patch(ctx, inode.ID(), map[string]any{
			"last_modified_at": args.ModifiedAt,
			"size":             newSize,
		})
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to Patch: %w", err))
		}

		inodeID = inode.Parent()
	}

	return nil
}
