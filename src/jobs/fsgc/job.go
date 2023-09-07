package fsgc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

const (
	gcBatchSize = 10
	jobName     = "fsgc"
)

type Job struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
	log     *slog.Logger
	cancel  context.CancelFunc
	quit    chan struct{}
}

func NewJob(inodes inodes.Service, files files.Service, folders folders.Service, tools tools.Tools) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{inodes, files, folders, logger, nil, make(chan struct{})}
}

func (j *Job) Run(ctx context.Context) error {
	j.log.DebugContext(ctx, "start job")
	for {
		toDelete, err := j.inodes.GetAllDeleted(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetAllDeleted: %w", err)
		}

		for _, inode := range toDelete {
			err = j.deleteINode(ctx, &inode)
			if err != nil {
				return fmt.Errorf("failed to delete inode %q: %w", inode.ID(), err)
			}

			j.log.DebugContext(ctx, "inode successfully removed", slog.String("inode", string(inode.ID())))
		}

		if len(toDelete) < gcBatchSize {
			j.log.DebugContext(ctx, "end job")
			return nil
		}
	}
}

func (j *Job) deleteDirINode(ctx context.Context, inode *inodes.INode) error {
	for {
		childs, err := j.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     inode.ID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: gcBatchSize})
		if err != nil {
			return fmt.Errorf("failed to Readdir: %w", err)
		}

		for _, child := range childs {
			err = j.deleteINode(ctx, &child)
			if err != nil {
				return fmt.Errorf("failed to deleteINode %q: %w", child.ID(), err)
			}
		}

		if len(childs) < gcBatchSize {
			break
		}
	}

	err := j.inodes.HardDelete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	return nil
}

func (j *Job) deleteINode(ctx context.Context, inode *inodes.INode) error {
	if inode.Mode().IsDir() {
		return j.deleteDirINode(ctx, inode)
	}

	// For the file we have several steps:
	//
	// - Reduce all the concerned folder size
	// - Remove the corresponding file
	// - Remove the corresponding inode
	rootInode, err := j.inodes.GetINodeRoot(ctx, inode)
	if err != nil {
		return fmt.Errorf("failed to GetINodeRoot: %w", err)
	}

	folders, err := j.folders.GetAllFoldersWithRoot(ctx, rootInode.ID(), nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllFoldersWithRoot: %w", err)
	}

	for _, folder := range folders {
		_, err = j.folders.RegisterDeletion(ctx, folder.ID(), uint64(inode.Size()))
		if err != nil {
			return fmt.Errorf("failed to RegisterDeletion: %w", err)
		}
	}

	err = j.files.Delete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to remove the file %q: %w", inode.ID(), err)
	}

	err = j.inodes.HardDelete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to HardDelete: %w", err)
	}

	return nil
}
