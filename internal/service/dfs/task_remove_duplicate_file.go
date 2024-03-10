package dfs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSRemoveDuplicateFilesRunner struct {
	storage   storage
	files     files.Service
	scheduler scheduler.Service
}

func NewFSRemoveDuplicateFileRunner(storage storage, files files.Service, scheduler scheduler.Service) *FSRemoveDuplicateFilesRunner {
	return &FSRemoveDuplicateFilesRunner{storage, files, scheduler}
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
	duplicates, err := r.storage.GetAllInodesWithFileID(ctx, args.DuplicateFileID)
	if err != nil {
		return fmt.Errorf("failed to GetAllInodesWithFileID: %w", err)
	}

	existingFileMeta, err := r.files.GetMetadata(ctx, args.ExistingFileID)
	if err != nil {
		return fmt.Errorf("failed to get the file: %w", err)
	}

	for _, duplicate := range duplicates {
		err := r.storage.Patch(ctx, duplicate.ID(), map[string]any{
			"file_id": existingFileMeta.ID(),
		})
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to Patch the inode: %w", err))
		}

		// XXX:MULTI-WRITE
		err = r.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
			INode:      duplicate.ID(),
			ModifiedAt: duplicate.LastModifiedAt(),
		})
		if err != nil {
			return fmt.Errorf("failed to scheduler the fs-refresh-size task: %w", err)
		}
	}

	err = r.files.Delete(ctx, args.DuplicateFileID)
	if err != nil {
		return fmt.Errorf("failed to Delete the old file id: %w", err)
	}

	return nil
}
