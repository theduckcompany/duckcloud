package inodes

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io/fs"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	ErrInvalidParent = errors.New("invalid parent")
	ErrIsNotDir      = errors.New("not a directory")
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
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(tools tools.Tools, storage Storage) *INodeService {
	return &INodeService{storage, tools.Clock(), tools.UUID()}
}

func (s *INodeService) GetByID(ctx context.Context, inodeID uuid.UUID) (*INode, error) {
	res, err := s.storage.GetByID(ctx, inodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByID: %w", err)
	}

	if res == nil {
		return nil, nil
	}

	return res, nil
}

func (s *INodeService) MkdirAll(ctx context.Context, cmd *PathCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	var inode *INode
	err = s.walk(ctx, cmd, "mkdir", func(dir *INode, frag string, _ bool) error {
		nextDir, err := s.storage.GetByNameAndParent(ctx, frag, dir.ID())
		if err != nil {
			return fmt.Errorf("failed to GetByNameAndParent: %w", err)
		}

		if nextDir != nil && nextDir.IsDir() {
			return nil
		}

		if nextDir != nil && !nextDir.IsDir() {
			return ErrIsNotDir
		}

		inode, err = s.CreateDir(ctx, &PathCmd{
			Root:     dir.ID(),
			FullName: frag,
		})
		if err != nil {
			return fmt.Errorf("failed to CreateDir: %w", err)
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
		isDir:          true,
		createdAt:      now,
		lastModifiedAt: now,
	}

	err := s.storage.Save(ctx, &node)
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

	now := s.clock.Now()
	inode := INode{
		id:             s.uuid.New(),
		parent:         ptr.To(parent.ID()),
		isDir:          false,
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

func (s *INodeService) RegisterWrite(ctx context.Context, inode *INode, sizeWrite int, h hash.Hash) error {
	inode.lastModifiedAt = s.clock.Now()
	inode.size += uint64(sizeWrite)
	inode.checksum = hex.EncodeToString(h.Sum(nil))

	return s.storage.Patch(ctx, inode.ID(), map[string]any{
		"last_modified_at": inode.lastModifiedAt,
		"size":             inode.size,
		"checksum":         inode.checksum,
	})
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

	res, err := s.storage.GetAllChildrens(ctx, dir.ID(), paginateCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAllChildrens: %w", err)
	}

	return res, nil
}

func (s *INodeService) GetAllDeleted(ctx context.Context, limit int) ([]INode, error) {
	return s.storage.GetAllDeleted(ctx, limit)
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

	return s.storage.Patch(ctx, inode.ID(), map[string]any{"deleted_at": s.clock.Now()})
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
			isDir:          true,
			parent:         ptr.To(dir.ID()),
			name:           frag,
			lastModifiedAt: now,
			createdAt:      now,
		}

		res, err := s.storage.GetByNameAndParent(ctx, frag, dir.ID())
		if err != nil {
			return fmt.Errorf("failed to GetByNameAndParent: %w", err)
		}

		if res != nil {
			return fs.ErrExist
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

		inode, err = s.storage.GetByNameAndParent(ctx, frag, dir.ID())
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

// GetINodeRoot returns the Root folder for any given inode.
func (s *INodeService) GetINodeRoot(ctx context.Context, inode *INode) (*INode, error) {
	for {
		if inode.parent == nil {
			return inode, nil
		}

		parent, err := s.GetByID(ctx, *inode.Parent())
		if err != nil {
			return nil, fmt.Errorf("failed to GetByID: %w", err)
		}

		if parent == nil {
			return inode, nil
		}

		inode = parent
	}
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
	original := cmd.FullName
	fullname := path.Clean("/" + cmd.FullName)

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

		child, err := s.storage.GetByNameAndParent(ctx, frag, dir.ID())
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
