package inodes

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const NoParent = uuid.UUID("00000000-0000-0000-0000-00000000000")

type PathCmd struct {
	SpaceID uuid.UUID
	Path    string
}

// Validate the fields.
func (t PathCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.SpaceID, v.Required, is.UUIDv4),
		v.Field(&t.Path, v.Required, v.Length(1, 1024)),
	)
}

type CreateFileCmd struct {
	SpaceID    uuid.UUID
	Parent     uuid.UUID
	Name       string
	FileID     uuid.UUID
	UploadedAt time.Time
	Key        secret.Text
}

// Validate the fields.
func (t CreateFileCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Parent, v.Required, is.UUIDv4),
		v.Field(&t.Name, v.Required, v.Length(1, 255)),
		v.Field(&t.FileID, v.Required, is.UUIDv4),
	)
}

type INode struct {
	id             uuid.UUID
	parent         *uuid.UUID
	name           string
	size           uint64
	spaceID        uuid.UUID
	createdAt      time.Time
	lastModifiedAt time.Time
	fileID         *uuid.UUID
}

func (n *INode) ID() uuid.UUID             { return n.id }
func (n *INode) Parent() *uuid.UUID        { return n.parent }
func (n *INode) Name() string              { return n.name }
func (n *INode) Size() uint64              { return n.size }
func (n *INode) SpaceID() uuid.UUID        { return n.spaceID }
func (n *INode) CreatedAt() time.Time      { return n.createdAt }
func (n *INode) LastModifiedAt() time.Time { return n.lastModifiedAt }
func (n *INode) FileID() *uuid.UUID        { return n.fileID }
func (n *INode) IsDir() bool               { return n.fileID == nil }
