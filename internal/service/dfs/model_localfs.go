package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type LocalFS struct {
	inodes    inodes.Service
	files     files.Service
	spaceID   uuid.UUID
	scheduler scheduler.Service
	clock     clock.Clock
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	spaceID uuid.UUID,
	tasks scheduler.Service,
	tools tools.Tools,
) *LocalFS {
	return &LocalFS{inodes, files, spaceID, tasks, tools.Clock()}
}

func (s *LocalFS) SpaceID() uuid.UUID {
	return s.spaceID
}

func (s *LocalFS) ListDir(ctx context.Context, path string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	path = CleanPath(path)

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    path,
	})
	if errors.Is(err, errs.ErrNotFound) {
		return nil, errs.NotFound(err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to Get inode: %w", err)
	}

	return s.inodes.Readdir(ctx, dir, cmd)
}

func (s *LocalFS) CreateDir(ctx context.Context, dirPath string) (*INode, error) {
	dirPath = CleanPath(dirPath)

	inode, err := s.inodes.MkdirAll(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    dirPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to MkdirAll: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) Remove(ctx context.Context, path string) error {
	path = CleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    path,
	})
	if errors.Is(err, errs.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to Get inode: %w", err)
	}

	err = s.inodes.Remove(ctx, res)
	if err != nil {
		return fmt.Errorf("failed to Remove: %w", err)
	}

	return nil
}

func (s *LocalFS) Move(ctx context.Context, oldPath, newPath string) error {
	sourceINode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    CleanPath(oldPath),
	})
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	err = s.scheduler.RegisterFSMoveTask(ctx, &scheduler.FSMoveArgs{
		SpaceID:     s.spaceID,
		SourceInode: sourceINode.ID(),
		TargetPath:  newPath,
		MovedAt:     s.clock.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to save the task: %w", err)
	}

	return nil
}

func (s *LocalFS) Get(ctx context.Context, path string) (*INode, error) {
	path = CleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return res, nil
}

func (s *LocalFS) Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error) {
	inode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	fileID := inode.FileID()
	if fileID == nil {
		return nil, files.ErrInodeNotAFile
	}

	fileMeta, err := s.files.GetMetadata(ctx, *fileID)
	if err != nil {
		return nil, err
	}

	fileReader, err := s.files.Download(ctx, fileMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to Open file %q: %w", inode.ID(), err)
	}

	return fileReader, nil
}

func (s *LocalFS) Upload(ctx context.Context, filePath string, w io.Reader) error {
	filePath = CleanPath(filePath)

	dirPath, fileName := path.Split(filePath)

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		SpaceID: s.spaceID,
		Path:    dirPath,
	})
	if err != nil {
		return fmt.Errorf("failed to get the dir: %w", err)
	}

	fileID, err := s.files.Upload(ctx, w)
	if err != nil {
		return fmt.Errorf("failed to Create file: %w", err)
	}

	ctx = context.WithoutCancel(ctx)
	now := s.clock.Now()

	// XXX:MULTI-WRITE
	//
	inode, err := s.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent:     dir.ID(),
		Name:       fileName,
		FileID:     fileID,
		UploadedAt: now,
	})
	if err != nil {
		return fmt.Errorf("failed to inodes.CreateFile: %w", err)
	}

	err = s.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
		INode:      inode.ID(),
		ModifiedAt: now,
	})
	if err != nil {
		return fmt.Errorf("failed to register the fs-refresh-size task: %w", err)
	}

	return nil
}
