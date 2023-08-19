package inodes

import (
	"io/fs"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const NoParent = uuid.UUID("00000000-0000-0000-0000-00000000000")

type PathCmd struct {
	Root     uuid.UUID
	UserID   uuid.UUID
	FullName string
}

// Validate the fields.
func (t PathCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Root, v.Required, is.UUIDv4),
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.FullName, v.Required, v.Length(1, 1024)),
	)
}

type CreateFileCmd struct {
	Parent uuid.UUID
	UserID uuid.UUID
	Name   string
	Mode   fs.FileMode
}

// Validate the fields.
func (t CreateFileCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Parent, v.Required, is.UUIDv4),
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.Name, v.Required, v.Length(1, 255)),
	)
}

type INode struct {
	id             uuid.UUID
	userID         uuid.UUID
	parent         uuid.UUID
	mode           fs.FileMode
	name           string
	size           int64
	createdAt      time.Time
	lastModifiedAt time.Time
}

func (n *INode) ID() uuid.UUID             { return n.id }
func (n *INode) UserID() uuid.UUID         { return n.userID }
func (n *INode) Parent() uuid.UUID         { return n.parent }
func (n *INode) Name() string              { return n.name }
func (n *INode) Size() int64               { return n.size }
func (n *INode) Mode() fs.FileMode         { return n.mode }
func (n *INode) ModTime() time.Time        { return n.lastModifiedAt }
func (n *INode) CreatedAt() time.Time      { return n.createdAt }
func (n *INode) LastModifiedAt() time.Time { return n.lastModifiedAt }
func (n *INode) IsDir() bool               { return n.mode.IsDir() }
func (n *INode) Sys() any                  { return nil }
