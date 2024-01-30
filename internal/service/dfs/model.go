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
	Space *spaces.Space
	Path  string
}

func (t PathCmd) Equal(p PathCmd) bool {
	return t.Space == p.Space && t.Path == p.Path
}

func (t PathCmd) String() string {
	return string(t.Space.ID()) + ":" + t.Path
}

// Contains returns true if the the arg `p` point to the same element
// or an element contained by `t`.
//
// "/foo".Contains("/foo/") -> true
// "/foo/bar".Contains("/foo") -> false
// "/foo".Contains("/foo/bar") -> true
func (t PathCmd) Contains(p PathCmd) bool {
	if t.Space != p.Space {
		return false
	}

	return strings.Contains(CleanPath(p.Path), CleanPath(t.Path))
}

// Validate the fields.
func (t PathCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Space, v.Required),
		v.Field(&t.Path, v.Required, v.Length(1, 1024)),
	)
}

type UploadCmd struct {
	Space      *spaces.Space
	FilePath   string
	Content    io.Reader
	UploadedBy *users.User
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
	Space     *spaces.Space
	FilePath  string
	CreatedBy *users.User
}

func (t CreateDirCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Space, v.Required, v.NotNil),
		v.Field(&t.FilePath, v.Required, v.Length(1, 255)),
		v.Field(&t.CreatedBy, v.Required, v.NotNil),
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
	id             uuid.UUID
	parent         *uuid.UUID
	name           string
	spaceID        uuid.UUID
	size           uint64
	createdAt      time.Time
	createdBy      uuid.UUID
	lastModifiedAt time.Time
	fileID         *uuid.UUID
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
