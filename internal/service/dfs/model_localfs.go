package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, dir *INode) error
	GetByID(ctx context.Context, id uuid.UUID) (*INode, error)
	GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error)
	GetAllChildrens(ctx context.Context, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	GetAllDeleted(ctx context.Context, limit int) ([]INode, error)
	GetDeleted(ctx context.Context, id uuid.UUID) (*INode, error)
	Patch(ctx context.Context, inode uuid.UUID, fields map[string]any) error
	GetSumChildsSize(ctx context.Context, parent uuid.UUID) (uint64, error)
	GetAllInodesWithFileID(ctx context.Context, fileID uuid.UUID) ([]INode, error)
	GetSpaceRoot(ctx context.Context, spaceID uuid.UUID) (*INode, error)
}

type LocalFS struct {
	storage   Storage
	inodes    inodes.Service
	files     files.Service
	space     *spaces.Space
	spaces    spaces.Service
	scheduler scheduler.Service
	clock     clock.Clock
	uuid      uuid.Service
}

func newLocalFS(
	storage Storage,
	inodes inodes.Service,
	files files.Service,
	space *spaces.Space,
	spaces spaces.Service,
	tasks scheduler.Service,
	tools tools.Tools,
) *LocalFS {
	return &LocalFS{storage, inodes, files, space, spaces, tasks, tools.Clock(), tools.UUID()}
}

func (s *LocalFS) Space() *spaces.Space {
	return s.space
}

func (s *LocalFS) ListDir(ctx context.Context, path string, cmd *storage.PaginateCmd) ([]INode, error) {
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

	if !dir.IsDir() {
		return nil, errs.BadRequest(ErrIsNotDir)
	}

	res, err := s.storage.GetAllChildrens(ctx, dir.ID(), cmd)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetAllChildrens: %w", err))
	}

	return res, nil
}

func (s *LocalFS) Rename(ctx context.Context, inode *INode, newName string) (*INode, error) {
	if newName == "" {
		return nil, errs.Validation(errors.New("can't be empty"))
	}

	newName, err := s.findUniqueName(ctx, inode, newName)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()

	newINode := *inode
	newINode.name = newName
	newINode.lastModifiedAt = now

	err = s.storage.Patch(ctx, inode.ID(), map[string]any{
		"name":             newName,
		"last_modified_at": now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Patch: %w", err)
	}

	return &newINode, err
}

func (s *LocalFS) findUniqueName(ctx context.Context, inode *INode, newName string) (string, error) {
	if inode.Parent() == nil {
		return "", errs.Validation(errors.New("can't rename the root"))
	}

	name := newName
	loop := 0
	ext := path.Ext(newName)
	base := strings.TrimRight(newName, ext)
	for {
		if loop > 0 {
			name = fmt.Sprintf("%s (%d)%s", base, loop, ext)
		}

		_, err := s.inodes.GetByNameAndParent(ctx, name, *inode.Parent())
		if errors.Is(err, errs.ErrNotFound) {
			return name, nil
		}

		if err != nil {
			return "", errs.Internal(fmt.Errorf("failed to check if the name is already taken: %w", err))
		}

		loop++
	}
}

func (s *LocalFS) CreateDir(ctx context.Context, cmd *CreateDirCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	var inode *INode
	currentPath := "/"
	err = s.walk(ctx, &PathCmd{
		Space: s.space,
		Path:  CleanPath(cmd.FilePath),
	}, "mkdir", func(dir *INode, frag string, _ bool) error {
		currentPath = path.Join(currentPath, dir.Name())

		if frag == "" {
			inode = dir
			return nil
		}

		nextDir, err := s.storage.GetByNameAndParent(ctx, frag, dir.ID())
		if err != nil && !errors.Is(err, errNotFound) {
			return errs.Internal(fmt.Errorf("failed to GetByNameAndParent: %w", err))
		}

		if nextDir != nil && nextDir.IsDir() {
			inode = nextDir
			return nil
		}

		if nextDir != nil && !nextDir.IsDir() {
			return errs.BadRequest(ErrIsNotDir)
		}

		// XXX:MULTI-WRITE
		//
		// This function is idempotent so there isn't a real issue here. Worst case
		// senario only some spaces are recreated but a new call would create them.
		inode, err = s.createDir(ctx, cmd.CreatedBy, dir, frag)
		if err != nil {
			return fmt.Errorf("failed to createDir %q: %w", path.Join(currentPath, frag), err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (s *LocalFS) Remove(ctx context.Context, path string) error {
	path = CleanPath(path)

	inode, err := s.Get(ctx, path)
	if errors.Is(err, errs.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to Get inode: %w", err)
	}

	now := s.clock.Now()
	err = s.storage.Patch(ctx, inode.ID(), map[string]any{
		"deleted_at":       now,
		"last_modified_at": now,
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Patch: %w", err))
	}

	if inode.parent != nil {
		err = s.scheduler.RegisterFSRefreshSizeTask(ctx, &scheduler.FSRefreshSizeArg{
			INode:      *inode.Parent(),
			ModifiedAt: now,
		})
		if err != nil {
			return fmt.Errorf("failed to schedule the fs-refresh-size task: %w", err)
		}
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

func (s *LocalFS) Get(ctx context.Context, pathStr string) (*INode, error) {
	cmd := PathCmd{
		Space: s.space,
		Path:  CleanPath(pathStr),
	}

	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	var inode *INode
	currentPath := "/"
	err = s.walk(ctx, &cmd, "open", func(dir *INode, frag string, final bool) error {
		currentPath = path.Join(currentPath, dir.Name())
		if !final {
			return nil
		}

		if frag == "" {
			inode = dir
			return nil
		}

		inode, err = s.storage.GetByNameAndParent(ctx, frag, dir.ID())
		if errors.Is(err, errNotFound) {
			return errs.NotFound(fmt.Errorf("%q doesn't have a child named %q", currentPath, frag))
		}

		if err != nil {
			return errs.Internal(fmt.Errorf("failed to fetch a file by name and parent: %w", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (s *LocalFS) Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error) {
	inode, err := s.Get(ctx, filePath)
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
		Parent:     dir,
		Name:       fileName,
		File:       fileID,
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

func (s *LocalFS) createDir(ctx context.Context, createdBy *users.User, parent *INode, name string) (*INode, error) {
	if !parent.IsDir() {
		return nil, errs.BadRequest(ErrIsNotDir)
	}

	res, err := s.storage.GetByNameAndParent(ctx, name, parent.ID())
	if err != nil && !errors.Is(err, errNotFound) {
		return nil, fmt.Errorf("failed to GetByNameAndParent: %w", err)
	}

	if res != nil {
		return nil, errs.BadRequest(ErrAlreadyExists)
	}

	now := s.clock.Now()
	newDir := INode{
		id:             s.uuid.New(),
		parent:         ptr.To(parent.ID()),
		name:           name,
		spaceID:        parent.SpaceID(),
		size:           0,
		createdAt:      now,
		createdBy:      createdBy.ID(),
		lastModifiedAt: now,
		fileID:         nil,
	}

	err = s.storage.Save(ctx, &newDir)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to save into the storage: %w", err))
	}

	return &newDir, nil
}

// walk walks the directory tree for the fullname, calling f at each step. If f
// returns an error, the walk will be aborted and return that same error.
//
// dir is the directory at that step, frag is the name fragment, and final is
// whether it is the final step. For example, walking "./foo/bar/x" will result
// in 3 calls to f:
//   - "/", "foo", false
//   - "/foo/", "bar", false
//   - "/foo/bar/", "x", true
//
// The frag argument will be empty only if dir is the root node and the walk
// ends at that root node.
func (s *LocalFS) walk(ctx context.Context, cmd *PathCmd, op string, f func(dir *INode, frag string, final bool) error) error {
	fullname := path.Clean("/" + cmd.Path)

	// Strip any leading "/"s to make fullname a relative path, as the walk
	// starts at fs.root.
	if fullname[0] == '/' {
		fullname = fullname[1:]
	}

	dir, err := s.storage.GetSpaceRoot(ctx, cmd.Space.ID())
	if errors.Is(err, errNotFound) {
		return errs.NotFound(ErrInvalidRoot, "failed to fetch the root dir for space %q", cmd.Space.Name())
	}
	rootFS := dir

	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetByID the root: %w", err))
	}

	for {
		frag, remaining := fullname, ""
		i := strings.IndexRune(fullname, '/')
		final := i < 0

		if !final {
			frag, remaining = fullname[:i], fullname[i+1:]
		}

		if frag == "" && dir.ID() != rootFS.id {
			panic("webdav: empty path fragment for a clean path")
		}

		if err := f(dir, frag, final); err != nil {
			return err
		}
		if final {
			break
		}

		child, err := s.storage.GetByNameAndParent(ctx, frag, dir.ID())
		if errors.Is(err, errNotFound) {
			return errs.NotFound(ErrInvalidPath)
		}
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to get child %q from %q: %w", frag, remaining, err))
		}

		if !child.IsDir() {
			return errs.BadRequest(fmt.Errorf("%s: %w", path.Join(remaining, frag), ErrIsNotDir))
		}
		dir, fullname = child, remaining
	}

	return nil
}
