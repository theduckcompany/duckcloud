package dfs

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/files"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
)

type FileUploadTaskRunner struct {
	folders folders.Service
	files   files.Service
	inodes  inodes.Service
}

func NewFileUploadTaskRunner(folders folders.Service, files files.Service, inodes inodes.Service) *FileUploadTaskRunner {
	return &FileUploadTaskRunner{folders, files, inodes}
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
	folder, err := r.folders.GetByID(ctx, args.FolderID)
	if err != nil {
		return fmt.Errorf("failed to get the folder: %w", err)
	}

	dirPath, fileName := path.Split(args.Path)

	dir, err := r.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Folder: folder,
		Path:   dirPath,
	})
	if err != nil {
		return fmt.Errorf("failed to Get the dir: %w", err)
	}

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

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to reset the file offset: %w", err)
	}

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return fmt.Errorf("failed to detete the mimetype: %w", err)
	}

	inode, err := r.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent:     dir.ID(),
		Name:       fileName,
		Mime:       mime.String(),
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
	err = r.inodes.RegisterWrite(ctx, inode, uint64(written), inode.LastModifiedAt())
	if err != nil {
		return fmt.Errorf("failed to RegisterWrite: %w", err)
	}

	return nil
}
