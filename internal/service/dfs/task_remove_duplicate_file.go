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
	duplicates, err := r.inodes.GetAllInodesWithFileID(ctx, args.DuplicateFileID)
	if err != nil {
		return fmt.Errorf("failed to GetAllInodesWithFileID: %w", err)
	}

	existingFileMeta, err := r.files.GetMetadata(ctx, args.ExistingFileID)
	if err != nil {
		return fmt.Errorf("failed to get the file: %w", err)
	}

	for _, duplicate := range duplicates {
		_, err := r.inodes.PatchFileID(ctx, &duplicate, existingFileMeta.ID())
		if err != nil {
			return fmt.Errorf("failed to update the file id: %w", err)
		}
	}

	err = r.files.Delete(ctx, args.DuplicateFileID)
	if err != nil {
		return fmt.Errorf("failed to Delete the old file id: %w", err)
	}

	return nil
}
