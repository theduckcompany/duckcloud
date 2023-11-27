package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes    inodes.Service
	files     files.Service
	space     *spaces.Space
	spaces    spaces.Service
	scheduler scheduler.Service
	clock     clock.Clock
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	space *spaces.Space,
	spaces spaces.Service,
	tasks scheduler.Service,
	tools tools.Tools,
) *LocalFS {
	return &LocalFS{inodes, files, space, spaces, tasks, tools.Clock()}
}

func (s *LocalFS) Space() *spaces.Space {
	return s.space
}

func (s *LocalFS) ListDir(ctx context.Context, path string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	path = CleanPath(path)

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  path,
	})
	if errors.Is(err, errs.ErrNotFound) {
		return nil, errs.NotFound(err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to Get inode: %w", err)
	}

	return s.inodes.Readdir(ctx, dir, cmd)
}

func (s *LocalFS) CreateDir(ctx context.Context, cmd *CreateDirCmd) (*INode, error) {
	dirPath := CleanPath(cmd.FilePath)

	inode, err := s.inodes.MkdirAll(ctx, cmd.CreatedBy, &inodes.PathCmd{
		Space: s.space,
		Path:  dirPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to MkdirAll: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) Remove(ctx context.Context, path string) error {
	path = CleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  path,
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

func (s *LocalFS) Move(ctx context.Context, cmd *MoveCmd) error {
	err := cmd.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	sourceINode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  CleanPath(cmd.SrcPath),
	})
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	err = s.scheduler.RegisterFSMoveTask(ctx, &scheduler.FSMoveArgs{
		SpaceID:     s.space.ID(),
		SourceInode: sourceINode.ID(),
		TargetPath:  cmd.NewPath,
		MovedAt:     s.clock.Now(),
		MovedBy:     cmd.MovedBy.ID(),
	})
	if err != nil {
		return fmt.Errorf("failed to save the task: %w", err)
	}

	return nil
}

func (s *LocalFS) Get(ctx context.Context, path string) (*INode, error) {
	path = CleanPath(path)

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return res, nil
}

func (s *LocalFS) Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error) {
	inode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  filePath,
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

func (s *LocalFS) Upload(ctx context.Context, cmd *UploadCmd) error {
	err := cmd.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	filePath := CleanPath(cmd.FilePath)

	dirPath, fileName := path.Split(filePath)

	dir, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Space: s.space,
		Path:  dirPath,
	})
	if err != nil {
		return fmt.Errorf("failed to get the dir: %w", err)
	}

	fileID, err := s.files.Upload(ctx, cmd.Content)
	if err != nil {
		return fmt.Errorf("failed to Create file: %w", err)
	}

	ctx = context.WithoutCancel(ctx)
	now := s.clock.Now()

	// XXX:MULTI-WRITE
	//
	inode, err := s.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Space:      s.space,
		Parent:     dir.ID(),
		Name:       fileName,
		FileID:     fileID,
		UploadedAt: now,
		UploadedBy: cmd.UploadedBy,
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
