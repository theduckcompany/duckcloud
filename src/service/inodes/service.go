package inodes

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	ErrInvalidParent      = errors.New("invalid parent")
	ErrAlreadyBootstraped = errors.New("this user is already bootstraped")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, dir *INode) error
	UpdateModifiedSizeAndDirty(ctx context.Context, inode *INode) error
	GetByID(ctx context.Context, id uuid.UUID) (*INode, error)
	CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error)
	GetByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parent uuid.UUID) (*INode, error)
	GetAllChildrens(ctx context.Context, userID, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error)
	Delete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	GetDeletedINodes(ctx context.Context, limit int) ([]INode, error)
}

type INodeService struct {
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(tools tools.Tools, storage Storage) *INodeService {
	return &INodeService{storage, tools.Clock(), tools.UUID()}
}

func (s *INodeService) BootstrapUser(ctx context.Context, userID uuid.UUID) (*INode, error) {
	nb, err := s.storage.CountUserINodes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count the number of inodes: %w", err)
	}

	if nb > 0 {
		return nil, errs.BadRequest(ErrAlreadyBootstraped, "user alread bootstraped")
	}

	now := s.clock.Now()

	node := INode{
		id:             s.uuid.New(),
		userID:         userID,
		parent:         NoParent,
		mode:           0o660 | fs.ModeDir,
		createdAt:      now,
		lastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}

func (s *INodeService) CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	parent, err := s.storage.GetByID(ctx, cmd.Parent)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByID: %w", err)
	}

	if parent == nil {
		return nil, fmt.Errorf("%w: parent %q not found", ErrInvalidParent, cmd.Parent)
	}

	if parent.UserID() != cmd.UserID {
		return nil, fmt.Errorf("%w: parent %q is owned by someone else", ErrInvalidParent, cmd.Parent)
	}

	now := s.clock.Now()
	inode := INode{
		id:             s.uuid.New(),
		parent:         parent.ID(),
		userID:         cmd.UserID,
		mode:           cmd.Mode,
		size:           0,
		name:           cmd.Name,
		createdAt:      now,
		lastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &inode)
	if err != nil {
		return nil, fmt.Errorf("failed to Save: %w", err)
	}

	return &inode, nil
}

func (s *INodeService) RegisterWrite(ctx context.Context, inode *INode, sizeWrite int) error {
	inode.lastModifiedAt = time.Now()
	inode.size += int64(sizeWrite)

	return s.storage.UpdateModifiedSizeAndDirty(ctx, inode)
}

func (s *INodeService) Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	dir, err := s.Get(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", cmd.FullName, err)
	}

	res, err := s.storage.GetAllChildrens(ctx, cmd.UserID, dir.ID(), paginateCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllChildrens: %w", err)
	}

	return res, nil
}

func (s *INodeService) GetDeletedINodes(ctx context.Context, limit int) ([]INode, error) {
	return s.storage.GetDeletedINodes(ctx, limit)
}

func (s *INodeService) HardDelete(ctx context.Context, inode uuid.UUID) error {
	return s.storage.HardDelete(ctx, inode)
}

func (s *INodeService) RemoveAll(ctx context.Context, cmd *PathCmd) error {
	err := cmd.Validate()
	if err != nil {
		return errs.ValidationError(err)
	}

	inode, err := s.Get(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to open the inode: %w", err)
	}

	if inode == nil {
		return nil
	}

	err = s.storage.Delete(ctx, inode.ID())
	if err != nil {
		return fmt.Errorf("failed to soft delete the inode %q: %w", inode.ID(), err)
	}

	return nil
}

func (s *INodeService) CreateDir(ctx context.Context, cmd *PathCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	var inode *INode
	err = s.walk(ctx, cmd, "mkdir", func(dir *INode, frag string, final bool) error {
		if !final {
			return nil
		}

		now := s.clock.Now()

		inode = &INode{
			id:             s.uuid.New(),
			userID:         cmd.UserID,
			mode:           0o660 | fs.ModeDir,
			parent:         dir.ID(),
			name:           frag,
			lastModifiedAt: now,
			createdAt:      now,
		}

		err = s.storage.Save(ctx, inode)
		if err != nil {
			return fmt.Errorf("failed to save into the storage: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (s *INodeService) Get(ctx context.Context, cmd *PathCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	var inode *INode
	err = s.walk(ctx, cmd, "open", func(dir *INode, frag string, final bool) error {
		if !final {
			return nil
		}

		if frag == "" {
			inode = dir
			return nil
		}

		inode, err = s.storage.GetByNameAndParent(ctx, cmd.UserID, frag, dir.ID())
		if err != nil {
			return fmt.Errorf("failed to fetch a file by name and parent: %w", err)
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
// whether it is the final step. For example, walking "/foo/bar/x" will result
// in 3 calls to f:
//   - "/", "foo", false
//   - "/foo/", "bar", false
//   - "/foo/bar/", "x", true
//
// The frag argument will be empty only if dir is the root node and the walk
// ends at that root node.
func (s *INodeService) walk(ctx context.Context, cmd *PathCmd, op string, f func(dir *INode, frag string, final bool) error) error {
	original := cmd.FullName
	fullname := slashClean(cmd.FullName)

	// Strip any leading "/"s to make fullname a relative path, as the walk
	// starts at fs.root.
	if fullname[0] == '/' {
		fullname = fullname[1:]
	}

	dir, err := s.storage.GetByID(ctx, cmd.Root)
	if err != nil {
		return fmt.Errorf("failed to fetch the root dir %q", cmd.Root)
	}

	if dir == nil {
		return errs.NotFound(fmt.Errorf("root %q not found", cmd.Root), "root not found")
	}

	if dir.UserID() != cmd.UserID {
		return errs.NotFound(fmt.Errorf("dir %q is not owned by %q", cmd.Root, cmd.UserID), "access denied")
	}

	for {
		frag, remaining := fullname, ""
		i := strings.IndexRune(fullname, '/')
		final := i < 0

		if !final {
			frag, remaining = fullname[:i], fullname[i+1:]
		}

		if frag == "" && dir.ID() != cmd.Root {
			panic("webdav: empty path fragment for a clean path")
		}

		if err := f(dir, frag, final); err != nil {
			return &fs.PathError{
				Op:   op,
				Path: original,
				Err:  err,
			}
		}
		if final {
			break
		}

		child, err := s.storage.GetByNameAndParent(ctx, cmd.UserID, frag, dir.ID())
		if err != nil {
			return fmt.Errorf("failed to get child %q from %q", frag, remaining)
		}

		if child == nil {
			return &fs.PathError{
				Op:   op,
				Path: original,
				Err:  fs.ErrNotExist,
			}
		}

		if !child.IsDir() {
			return &fs.PathError{
				Op:   op,
				Path: original,
				Err:  fs.ErrInvalid,
			}
		}
		dir, fullname = child, remaining
	}

	return nil
}

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
