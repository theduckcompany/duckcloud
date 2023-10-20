package dfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSMoveTaskRunner struct {
	inodes  inodes.Service
	folders folders.Service
}

func NewFSMoveTaskRunner(inodes inodes.Service, folders folders.Service) *FSMoveTaskRunner {
	return &FSMoveTaskRunner{inodes, folders}
}

func (r *FSMoveTaskRunner) Name() string { return "fs-move" }

func (r *FSMoveTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.FSMoveArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *FSMoveTaskRunner) RunArgs(ctx context.Context, args *scheduler.FSMoveArgs) error {
	folder, err := r.folders.GetByID(ctx, args.FolderID)
	if err != nil {
		return fmt.Errorf("failed to Get the folder: %w", err)
	}

	existingFile, err := r.inodes.Get(ctx, &inodes.PathCmd{
		Folder: folder,
		Path:   args.TargetPath,
	})
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("failed to check if a file already existed: %w", err)
	}

	oldNode, err := r.inodes.GetByID(ctx, args.SourceInode)
	if err != nil {
		return fmt.Errorf("failed to GetByID %q: %w", args.SourceInode, err)
	}

	dir, filename := path.Split(args.TargetPath)

	targetDir, err := r.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Folder: folder,
		Path:   dir,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch the source: %w", err)
	}

	if oldNode.Parent() != nil {
		parentOldNode, err := r.inodes.GetByID(ctx, *oldNode.Parent())
		if err != nil {
			return fmt.Errorf("failed to retrieve the old node parent: %w", err)
		}

		if parentOldNode.LastModifiedAt().Before(args.MovedAt) {
			// XXX:MULTI-WRITE
			//
			// This call is idemptotent and will be executed only once.
			err = r.inodes.RegisterWrite(ctx, oldNode, -oldNode.Size(), args.MovedAt)
			if err != nil {
				return fmt.Errorf("failed to remove the old inode file size: %w", err)
			}
		}
	}

	newNode, err := r.inodes.PatchMove(ctx, oldNode, targetDir, filename, args.MovedAt)
	if err != nil {
		return fmt.Errorf("failed to PatchMove: %w", err)
	}

	if existingFile != nil {
		// XXX:MULTI-WRITE
		//
		// During a move the old file should be removed. In case of error we can end's
		// with the old and the new file. This is not really dangerous as we don't loose
		// any data but both files will have the exact same name and this can be
		// problematic for the deletion for the manual example. We can't know which one
		// will be selected if we delete base on a path.
		//
		// TODO: Fix this with a commit system
		err = r.inodes.Remove(ctx, existingFile)
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to remove the old file: %w", err))
		}
	}

	if newNode.LastModifiedAt().Before(args.MovedAt) {
		err = r.inodes.RegisterWrite(ctx, newNode, newNode.Size(), args.MovedAt)
		if err != nil {
			return fmt.Errorf("failed to add the new inode file size: %w", err)
		}
	}

	return nil
}
