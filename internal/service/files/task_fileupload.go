package files

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
)

type FileUploadTaskRunner struct {
	files     Service
	storage   Storage
	scheduler scheduler.Service
}

func NewFileUploadTaskRunner(files Service, storage Storage, scheduler scheduler.Service) *FileUploadTaskRunner {
	return &FileUploadTaskRunner{files, storage, scheduler}
}

func (r *FileUploadTaskRunner) Name() string { return "file-upload" }

func (r *FileUploadTaskRunner) Run(ctx context.Context, rawArgs json.RawMessage) error {
	var args scheduler.FileUploadArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return r.RunArgs(ctx, &args)
}

func (r *FileUploadTaskRunner) RunArgs(ctx context.Context, args *scheduler.FileUploadArgs) error {
	file, err := r.files.Download(ctx, args.FileID)
	if err != nil {
		return fmt.Errorf("failed to files.Open: %w", err)
	}

	defer file.Close()

	hasher := sha256.New()
	written, err := io.Copy(hasher, file)
	if err != nil {
		return fmt.Errorf("failed to generate the hash: %w", err)
	}

	checksum := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	existingFile, err := r.storage.GetByChecksum(ctx, checksum)
	if err != nil && !errors.Is(err, errNotFound) {
		return fmt.Errorf("failed to GetByChecksum: %w", err)
	}

	if existingFile != nil {
		err = r.scheduler.RegisterFSRemoveDuplicateFile(ctx, &scheduler.FSRemoveDuplicateFileArgs{
			ExistingFileID:  existingFile.ID(),
			DuplicateFileID: args.FileID,
		})
		if err != nil {
			return fmt.Errorf("failed to register the fs-remove-duplicate-files task: %w", err)
		}

		return nil
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to reset the file offset: %w", err)
	}

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return fmt.Errorf("failed to detete the mimetype: %w", err)
	}

	err = r.storage.Save(ctx, &FileMeta{
		id:         args.FileID,
		size:       uint64(written),
		mimetype:   mime.String(),
		checksum:   checksum,
		uploadedAt: args.UploadedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to save the file meta: %w", err)
	}

	// // XXX:MULTI-WRITE
	// //
	// // This file have severa consecutive writes but they are all idempotent and the
	// // task is retried in case of error.
	err = r.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
		INode:      args.INodeID,
		ModifiedAt: args.UploadedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to register the fs-refresh-size task: %w", err)
	}

	return nil
}
