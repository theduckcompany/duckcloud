package inodes

import (
	"io/fs"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const NoParent = uuid.UUID("00000000-0000-0000-0000-00000000000")

type NodeType int

const (
	Directory NodeType = 0
	File      NodeType = 1
)

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
		v.Field(&t.FullName, v.Required, v.Length(1, 1024), is.PrintableASCII),
	)
}

type INode struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Parent         uuid.UUID
	Type           NodeType
	name           string
	CreatedAt      time.Time
	LastModifiedAt time.Time
}

func (n *INode) Name() string {
	return n.name
}

func (n *INode) Size() int64 {
	return 0
}

func (n *INode) Mode() fs.FileMode {
	return fs.ModeDir
}

func (n *INode) ModTime() time.Time {
	return n.LastModifiedAt
}

func (n *INode) IsDir() bool {
	return n.Type == Directory
}

func (d *INode) Sys() any {
	return nil
}
