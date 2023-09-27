package inodes

import (
	"io/fs"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const NoParent = uuid.UUID("00000000-0000-0000-0000-00000000000")

type PathCmd struct {
	Root     uuid.UUID
	FullName string
}

// Validate the fields.
func (t PathCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Root, v.Required, is.UUIDv4),
		v.Field(&t.FullName, v.Required, v.Length(1, 1024)),
	)
}

type CreateFileCmd struct {
	Parent uuid.UUID
	Name   string
}

// Validate the fields.
func (t CreateFileCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Parent, v.Required, is.UUIDv4),
		v.Field(&t.Name, v.Required, v.Length(1, 255)),
	)
}

type INode struct {
	id             uuid.UUID
	parent         *uuid.UUID
	name           string
	checksum       string
	size           uint64
	createdAt      time.Time
	lastModifiedAt time.Time
	fileID         *uuid.UUID
}

func (n *INode) ID() uuid.UUID             { return n.id }
func (n *INode) Parent() *uuid.UUID        { return n.parent }
func (n *INode) Name() string              { return n.name }
func (n *INode) Size() int64               { return int64(n.size) }
func (n *INode) USize() uint64             { return n.size }
func (n *INode) ModTime() time.Time        { return n.lastModifiedAt }
func (n *INode) CreatedAt() time.Time      { return n.createdAt }
func (n *INode) LastModifiedAt() time.Time { return n.lastModifiedAt }
func (n *INode) FileID() *uuid.UUID        { return n.fileID }
func (n *INode) IsDir() bool               { return n.fileID == nil }
func (n *INode) Checksum() string          { return n.checksum }
func (n *INode) Sys() any                  { return nil }

func (n *INode) Mode() fs.FileMode {
	if n.IsDir() {
		return 0o755 | fs.ModeDir
	}

	return 0o644 // Regular file
}
