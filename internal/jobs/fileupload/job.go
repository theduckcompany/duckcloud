package fileupload

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

const (
	batchSize = 10
	jobName   = "fileupload"
)

type Job struct {
	folders folders.Service
	uploads uploads.Service
	files   files.Service
	inodes  inodes.Service
	log     *slog.Logger
}

func NewJob(folders folders.Service, uploads uploads.Service, files files.Service, inodes inodes.Service, tools tools.Tools) *Job {
	logger := tools.Logger().With(slog.String("job", jobName))
	return &Job{folders, uploads, files, inodes, logger}
}

func (j *Job) Run(ctx context.Context) error {
	for {
		upload, err := j.uploads.GetOldest(ctx)
		if err != nil {
			return fmt.Errorf("failed to uploads.GetOldest: %w", err)
		}

		if upload == nil {
			return nil
		}

		file, err := j.files.Open(ctx, upload.FileID())
		if err != nil {
			return fmt.Errorf("failed to files.Open: %w", err)
		}

		defer file.Close()

		hasher := sha256.New()
		written, err := io.Copy(hasher, file)
		if err != nil {
			return fmt.Errorf("failed to generate the hash: %w", err)
		}

		inode, err := j.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
			Parent:     upload.Dir(),
			Name:       upload.FileName(),
			Size:       uint64(written),
			Checksum:   base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
			FileID:     upload.FileID(),
			UploadedAt: upload.UploadedAt(),
		})
		if err != nil {
			return fmt.Errorf("failed to inodes.CreateFile: %w", err)
		}

		parentID := inode.Parent()
		for {
			if parentID == nil {
				break
			}

			parent, err := j.inodes.GetByID(ctx, *parentID)
			if err != nil {
				return fmt.Errorf("failed to GetByID the parent: %w", err)
			}

			if !parent.LastModifiedAt().Equal(inode.LastModifiedAt()) {
				err = j.inodes.RegisterWrite(ctx, parent, written, inode.LastModifiedAt())
				if err != nil {
					return fmt.Errorf("failed to RegisterWrite: %w", err)
				}
			}

			parentID = parent.Parent()
		}

		err = j.uploads.Delete(ctx, upload)
		if err != nil {
			return fmt.Errorf("faile to delet the upload: %w", err)
		}
	}
}
