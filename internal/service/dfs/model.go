package dfs

import (
	io "io"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const NoParent = uuid.UUID("00000000-0000-0000-0000-00000000000")

type PathCmd struct {
	space *spaces.Space
	path  string
}

func NewPathCmd(space *spaces.Space, path string) *PathCmd {
	if space == nil {
		panic("NewPathCmd invoked with an nil space")
	}

	return &PathCmd{
		space: space,
		path:  CleanPath(path),
	}
}

func (t PathCmd) Path() string {
	return t.path
}

func (t PathCmd) Space() *spaces.Space {
	return t.space
}

func (t PathCmd) Equal(p PathCmd) bool {
	return t.space == p.space && t.path == p.path
}

func (t PathCmd) String() string {
	return string(t.space.ID()) + ":" + t.path
}

// Contains returns true if the the arg `p` point to the same element
// or an element contained by `t`.
//
// "/foo".Contains("/foo/") -> true
// "/foo/bar".Contains("/foo") -> false
// "/foo".Contains("/foo/bar") -> true
func (t PathCmd) Contains(p PathCmd) bool {
	if t.space != p.space {
		return false
	}

	return strings.Contains(CleanPath(p.path), CleanPath(t.path))
}

// Validate the fields.
func (t PathCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.space, v.Required),
		v.Field(&t.path, v.Required, v.Length(1, 1024)),
	)
}

type UploadCmd struct {
	Content    io.Reader
	Space      *spaces.Space
	UploadedBy *users.User
	FilePath   string
}

func (t UploadCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Space, v.Required, v.NotNil),
		v.Field(&t.FilePath, v.Required, v.Length(1, 255)),
		v.Field(&t.Content, v.Required),
		v.Field(&t.UploadedBy, v.Required, v.NotNil),
	)
}

type CreateDirCmd struct {
	Path      *PathCmd
	CreatedBy *users.User
}

func (t CreateDirCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Path, v.Required),
		v.Field(&t.CreatedBy, v.Required),
	)
}

type MoveCmd struct {
	Src     *PathCmd
	Dst     *PathCmd
	MovedBy *users.User
}

func (t MoveCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Src, v.Required, v.NotNil),
		v.Field(&t.Dst, v.Required, v.NotNil),
		v.Field(&t.MovedBy, v.Required, v.NotNil),
	)
}

type CreateRootDirCmd struct {
	CreatedBy *users.User
	Space     *spaces.Space
}

func (t CreateRootDirCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Space, v.Required, v.NotNil),
		v.Field(&t.CreatedBy, v.Required, v.NotNil),
	)
}

type INode struct {
	createdAt      time.Time
	lastModifiedAt time.Time
	parent         *uuid.UUID
	fileID         *uuid.UUID
	id             uuid.UUID
	name           string
	spaceID        uuid.UUID
	createdBy      uuid.UUID
	size           uint64
}

func (n INode) ID() uuid.UUID             { return n.id }
func (n INode) Parent() *uuid.UUID        { return n.parent }
func (n INode) Name() string              { return n.name }
func (n INode) SpaceID() uuid.UUID        { return n.spaceID }
func (n INode) Size() uint64              { return n.size }
func (n INode) CreatedAt() time.Time      { return n.createdAt }
func (n INode) CreatedBy() uuid.UUID      { return n.createdBy }
func (n INode) LastModifiedAt() time.Time { return n.lastModifiedAt }
func (n INode) FileID() *uuid.UUID        { return n.fileID }
func (n INode) IsDir() bool               { return n.fileID == nil }
