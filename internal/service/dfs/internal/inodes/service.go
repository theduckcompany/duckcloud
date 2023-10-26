package inodes

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrInvalidPath   = errors.New("invalid path")
	ErrInvalidRoot   = errors.New("invalid root")
	ErrInvalidParent = errors.New("invalid parent")
	ErrIsNotDir      = errors.New("not a directory")
	ErrIsADir        = errors.New("is a directory")
	ErrNotFound      = errors.New("inode not found")
	ErrAlreadyExists = errors.New("already exists")
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
}

type INodeService struct {
	scheduler scheduler.Service
	storage   Storage
	clock     clock.Clock
	uuid      uuid.Service
}

func NewService(scheduler scheduler.Service, tools tools.Tools, storage Storage) *INodeService {
	return &INodeService{scheduler, storage, tools.Clock(), tools.UUID()}
}

func (s *INodeService) GetByID(ctx context.Context, inodeID uuid.UUID) (*INode, error) {
	res, err := s.storage.GetByID(ctx, inodeID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(ErrNotFound)
	}
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	return res, nil
}

func (s *INodeService) MkdirAll(ctx context.Context, cmd *PathCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	var inode *INode
	currentPath := "/"
	err = s.walk(ctx, cmd, "mkdir", func(dir *INode, frag string, _ bool) error {
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
		// senario only some folders are recreated but a new call would create them.
		inode, err = s.CreateDir(ctx, dir, frag)
		if err != nil {
			return fmt.Errorf("failed to CreateDir %q: %w", path.Join(currentPath, frag), err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (s *INodeService) CreateRootDir(ctx context.Context) (*INode, error) {
	now := s.clock.Now()

	node := INode{
		id:             s.uuid.New(),
		parent:         nil,
		createdAt:      now,
		lastModifiedAt: now,
		fileID:         nil,
	}

	err := s.storage.Save(ctx, &node)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save: %w", err))
	}

	return &node, nil
}

func (s *INodeService) CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	parent, err := s.storage.GetByID(ctx, cmd.Parent)
	if errors.Is(err, errNotFound) {
		return nil, errs.BadRequest(fmt.Errorf("%w: parent %q not found", ErrInvalidParent, cmd.Parent))
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	inode := INode{
		id:             s.uuid.New(),
		parent:         ptr.To(parent.ID()),
		size:           cmd.Size,
		checksum:       cmd.Checksum,
		name:           cmd.Name,
		createdAt:      cmd.UploadedAt,
		lastModifiedAt: cmd.UploadedAt,
		fileID:         &cmd.FileID,
	}

	err = s.storage.Save(ctx, &inode)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save: %w", err))
	}

	return &inode, nil
}

// RegisterWrite will add/deduce the `sizeWrite` to all the given inode parents. The ModeTime will also
// be change for all the parents.
//
// This fonction contains several consecutive writes and is idempotent. In case of error you should call this
// function again. This keep the data coherent.
func (s *INodeService) RegisterWrite(ctx context.Context, inode *INode, sizeWrite int64, modeTime time.Time) error {
	var gerr error

	parentID := inode.Parent()
	for {
		if parentID == nil {
			break
		}

		parent, err := s.GetByID(ctx, *parentID)
		if err != nil {
			return fmt.Errorf("failed to GetByID the parent: %w", err)
		}

		if !parent.LastModifiedAt().Equal(modeTime) {
			parent.lastModifiedAt = modeTime
			if sizeWrite > 0 {
				parent.size += uint64(sizeWrite)
			} else {
				absVal := uint64(math.Abs(float64(sizeWrite)))
				if parent.size > absVal {
					parent.size -= absVal
				} else {
					parent.size = 0
				}
			}

			// XXX:MULTI-WRITE
			//
			// Those writes are done inside tasks and are indempotent so they should
			// be retried. In the worst case scenario we do not register a write size
			// and this is not dramatic.
			//
			// TODO: Create a job that will recalculate the correct size from time to time.
			err := s.storage.Patch(ctx, parent.ID(), map[string]any{
				"last_modified_at": parent.lastModifiedAt,
				"size":             parent.size,
			})
			if err != nil {
				gerr = fmt.Errorf("failed to Patch: %w", err)
			}
		}

		parentID = parent.Parent()
	}

	if gerr != nil {
		return errs.Internal(gerr)
	}

	return nil
}

func (s *INodeService) Readdir(ctx context.Context, dir *INode, paginateCmd *storage.PaginateCmd) ([]INode, error) {
	if !dir.IsDir() {
		return nil, errs.BadRequest(ErrIsNotDir)
	}

	res, err := s.storage.GetAllChildrens(ctx, dir.ID(), paginateCmd)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetAllChildrens: %w", err))
	}

	return res, nil
}

func (s *INodeService) GetAllDeleted(ctx context.Context, limit int) ([]INode, error) {
	res, err := s.storage.GetAllDeleted(ctx, limit)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *INodeService) HardDelete(ctx context.Context, inode uuid.UUID) error {
	err := s.storage.HardDelete(ctx, inode)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *INodeService) Remove(ctx context.Context, inode *INode) error {
	now := s.clock.Now()
	err := s.storage.Patch(ctx, inode.ID(), map[string]any{
		"deleted_at":       now,
		"last_modified_at": now,
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Patch: %w", err))
	}

	return nil
}

func (s *INodeService) GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error) {
	res, err := s.storage.GetByNameAndParent(ctx, name, parent)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *INodeService) PatchMove(ctx context.Context, source, parent *INode, newName string, modeTime time.Time) (*INode, error) {
	newFile := *source
	newFile.name = newName
	newFile.parent = &parent.id
	newFile.lastModifiedAt = modeTime

	err := s.storage.Patch(ctx, source.ID(), map[string]any{
		"parent":           *newFile.parent,
		"name":             newFile.name,
		"last_modified_at": newFile.lastModifiedAt,
	})
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to PatchMove the inode: %w", err))
	}

	return &newFile, nil
}

func (s *INodeService) CreateDir(ctx context.Context, parent *INode, name string) (*INode, error) {
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
		lastModifiedAt: now,
		createdAt:      now,
		fileID:         nil,
	}

	err = s.storage.Save(ctx, &newDir)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to save into the storage: %w", err))
	}

	return &newDir, nil
}

func (s *INodeService) Get(ctx context.Context, cmd *PathCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	var inode *INode
	currentPath := "/"
	err = s.walk(ctx, cmd, "open", func(dir *INode, frag string, final bool) error {
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
func (s *INodeService) walk(ctx context.Context, cmd *PathCmd, op string, f func(dir *INode, frag string, final bool) error) error {
	fullname := path.Clean("/" + cmd.Path)

	// Strip any leading "/"s to make fullname a relative path, as the walk
	// starts at fs.root.
	if fullname[0] == '/' {
		fullname = fullname[1:]
	}

	dir, err := s.storage.GetByID(ctx, cmd.Folder.RootFS())
	if errors.Is(err, errNotFound) {
		return errs.NotFound(ErrInvalidRoot, "failed to fetch the root dir for folder %q", cmd.Folder.Name())
	}

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

		if frag == "" && dir.ID() != cmd.Folder.RootFS() {
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
			return errs.Internal(fmt.Errorf("failed to get child %q from %q", frag, remaining))
		}

		if !child.IsDir() {
			return errs.BadRequest(fmt.Errorf("%s: %w", path.Join(remaining, frag), ErrIsNotDir))
		}
		dir, fullname = child, remaining
	}

	return nil
}
