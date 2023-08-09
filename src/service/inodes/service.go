package inodes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	ErrInvalidParent      = errors.New("invalid parent")
	ErrAlreadyBootstraped = errors.New("this user is already bootstraped")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, dir *INode) error
	GetByID(ctx context.Context, id uuid.UUID) (*INode, error)
	CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error)
	GetByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parent uuid.UUID) (*INode, error)
	GetAllChildrens(ctx context.Context, userID, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error)
	Remove(ctx context.Context, id uuid.UUID) error
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
		nodeType:       Directory,
		createdAt:      now,
		lastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}

func (s *INodeService) Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	dir, err := s.Open(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", cmd.FullName, err)
	}

	res, err := s.storage.GetAllChildrens(ctx, cmd.UserID, dir.ID(), paginateCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllChildrens: %w", err)
	}

	return res, nil
}

func (s *INodeService) Mkdir(ctx context.Context, cmd *PathCmd) (*INode, error) {
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
			nodeType:       Directory,
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

func (s *INodeService) Open(ctx context.Context, cmd *PathCmd) (*INode, error) {
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
			return &os.PathError{
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
			return &os.PathError{
				Op:   op,
				Path: original,
				Err:  os.ErrNotExist,
			}
		}

		if !child.IsDir() {
			return &os.PathError{
				Op:   op,
				Path: original,
				Err:  os.ErrInvalid,
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
