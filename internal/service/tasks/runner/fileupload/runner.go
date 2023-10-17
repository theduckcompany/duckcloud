package fileupload

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
)

type TaskRunner struct {
	folders folders.Service
	files   files.Service
	inodes  inodes.Service
}

func NewTaskRunner(folders folders.Service, files files.Service, inodes inodes.Service) *TaskRunner {
	return &TaskRunner{folders, files, inodes}
}

func (r *TaskRunner) Name() string { return model.FileUpload }

func (r *TaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.FileUploadArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *TaskRunner) RunArgs(ctx context.Context, args *scheduler.FileUploadArgs) error {
	file, err := r.files.Open(ctx, args.FileID)
	if err != nil {
		return fmt.Errorf("failed to files.Open: %w", err)
	}

	defer file.Close()

	hasher := sha256.New()
	written, err := io.Copy(hasher, file)
	if err != nil {
		return fmt.Errorf("failed to generate the hash: %w", err)
	}

	inode, err := r.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent:     args.Directory,
		Name:       args.FileName,
		Size:       uint64(written),
		Checksum:   base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
		FileID:     args.FileID,
		UploadedAt: args.UploadedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to inodes.CreateFile: %w", err)
	}

	// XXX:MULTI-WRITE
	//
	// This file have severa consecutive writes but they are all idempotent and the
	// task is retried in case of error.
	err = r.inodes.RegisterWrite(ctx, inode, written, inode.LastModifiedAt())
	if err != nil {
		return fmt.Errorf("failed to RegisterWrite: %w", err)
	}

	return nil
}
