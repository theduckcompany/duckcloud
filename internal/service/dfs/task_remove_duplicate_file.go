package dfs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSRemoveDuplicateFilesRunner struct {
	inodes    inodes.Service
	files     files.Service
	scheduler scheduler.Service
}

func NewFSRemoveDuplicateFileRunner(inodes inodes.Service, files files.Service, scheduler scheduler.Service) *FSRemoveDuplicateFilesRunner {
	return &FSRemoveDuplicateFilesRunner{inodes, files, scheduler}
}

func (r *FSRemoveDuplicateFilesRunner) Name() string { return "fs-remove-duplicate-file" }

func (r *FSRemoveDuplicateFilesRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.FSRemoveDuplicateFileArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to unmarshal the args: %w", err))
	}

	return r.RunArgs(ctx, &args)
}

func (r *FSRemoveDuplicateFilesRunner) RunArgs(ctx context.Context, args *scheduler.FSRemoveDuplicateFileArgs) error {
	inode, err := r.inodes.GetByID(ctx, args.INode)
	if err != nil {
		return fmt.Errorf("failed to get the inode: %w", err)
	}

	oldFileID := inode.FileID()

	fileMeta, err := r.files.GetMetadata(ctx, args.TargetFileID)
	if err != nil {
		return fmt.Errorf("failed to get the file: %w", err)
	}

	inode, err = r.inodes.PatchFileID(ctx, inode, fileMeta.ID())
	if err != nil {
		return fmt.Errorf("failed to update the file id: %w", err)
	}

	err = r.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
		INode:      inode.ID(),
		ModifiedAt: inode.LastModifiedAt(),
	})
	if err != nil {
		return fmt.Errorf("failed to scheduler the fs-refresh-size task: %w", err)
	}

	if oldFileID != nil {
		err = r.files.Delete(ctx, *oldFileID)
		if err != nil {
			return fmt.Errorf("failed to Delete the old file id: %w", err)
		}
	}

	return nil
}
