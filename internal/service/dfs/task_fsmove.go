package dfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type FSMoveTaskRunner struct {
	fs        Service
	storage   Storage
	spaces    spaces.Service
	users     users.Service
	scheduler scheduler.Service
}

func NewFSMoveTaskRunner(fs Service, storage Storage, spaces spaces.Service, users users.Service, scheduler scheduler.Service) *FSMoveTaskRunner {
	return &FSMoveTaskRunner{fs, storage, spaces, users, scheduler}
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
	space, err := r.spaces.GetByID(ctx, args.SpaceID)
	if err != nil {
		return fmt.Errorf("failed to Get the space: %w", err)
	}

	user, err := r.users.GetByID(ctx, args.MovedBy)
	if err != nil {
		return fmt.Errorf("failed to get the user: %w", err)
	}

	existingFile, err := r.fs.Get(ctx, NewPathCmd(space, args.TargetPath))
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("failed to check if a file already existed: %w", err)
	}

	oldNode, err := r.storage.GetByID(ctx, args.SourceInode)
	if err != nil {
		return fmt.Errorf("failed to GetByID %q: %w", args.SourceInode, err)
	}

	dir, filename := path.Split(args.TargetPath)

	targetDir, err := r.fs.CreateDir(ctx, &CreateDirCmd{
		Path:      NewPathCmd(space, dir),
		CreatedBy: user,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch the source: %w", err)
	}

	// XXX:MULTI-WRITE
	//
	//
	newNode := *oldNode
	newNode.name = filename
	newNode.parent = &targetDir.id
	newNode.lastModifiedAt = args.MovedAt

	err = r.storage.Patch(ctx, oldNode.ID(), map[string]any{
		"parent":           newNode.parent,
		"name":             newNode.name,
		"last_modified_at": newNode.lastModifiedAt,
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to PatchMove the inode: %w", err))
	}

	ctx = context.WithoutCancel(ctx)

	if existingFile != nil {
		// XXX:MULTI-WRITE
		//
		// During a move the old file should be removed. In case of error we can end's
		// with the old and the new file. This is not really dangerous as we don't loose
		// any data but both files will have the exact same name and this can be
		// problematic. We can't know which one will be selected if we delete based on a
		// path for example.
		//
		// TODO: Fix this with a commit system
		err = r.fs.removeINode(ctx, existingFile)
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to remove the old file: %w", err))
		}
	}

	if oldNode.Parent() != nil {
		err = r.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
			INode:      *oldNode.Parent(),
			ModifiedAt: args.MovedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to schedule the fs-refresh-size task for the old node: %w", err)
		}
	}

	if newNode.Parent() != nil {
		err = r.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
			INode:      *newNode.Parent(),
			ModifiedAt: args.MovedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to schedule the fs-refresh-size task for the new node: %w", err)
		}
	}

	return nil
}
