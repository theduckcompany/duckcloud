package scheduler

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FileUploadArgs struct {
	FolderID   uuid.UUID `json:"folder-id"`
	FileID     uuid.UUID `json:"file-id"`
	INodeID    uuid.UUID `json:"inode-id"`
	UploadedAt time.Time `json:"uploaded-at"`
}

func (a FileUploadArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.FolderID, v.Required, is.UUIDv4),
		v.Field(&a.INodeID, v.Required, is.UUIDv4),
		v.Field(&a.FileID, v.Required, is.UUIDv4),
		v.Field(&a.UploadedAt, v.Required),
	)
}

type FSMoveArgs struct {
	FolderID    uuid.UUID `json:"folder"`
	SourceInode uuid.UUID `json:"source-inode"`
	TargetPath  string    `json:"target-path"`
	MovedAt     time.Time `json:"moved-at"`
}

func (a FSMoveArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.FolderID, v.Required, is.UUIDv4),
		v.Field(&a.SourceInode, v.Required, is.UUIDv4),
		v.Field(&a.TargetPath, v.Required),
		v.Field(&a.MovedAt, v.Required),
	)
}

type FSGCArgs struct{}

func (a FSGCArgs) Validate() error {
	return v.ValidateStruct(&a)
}

type UserCreateArgs struct {
	UserID uuid.UUID `json:"user-id"`
}

func (a UserCreateArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.UserID, v.Required, is.UUIDv4),
	)
}

type UserDeleteArgs struct {
	UserID uuid.UUID `json:"user-id"`
}

func (a UserDeleteArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.UserID, v.Required, is.UUIDv4),
	)
}

type FSRefreshSizeArg struct {
	INode      uuid.UUID `json:"inode"`
	ModifiedAt time.Time `json:"modified_at"`
}

func (a FSRefreshSizeArg) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.INode, v.Required, is.UUIDv4),
		v.Field(&a.ModifiedAt, v.Required),
	)
}

type FSRemoveDuplicateFileArgs struct {
	INode        uuid.UUID `json:"inode"`
	TargetFileID uuid.UUID `json:"target-file-id"`
}

func (a FSRemoveDuplicateFileArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.INode, v.Required, is.UUIDv4),
		v.Field(&a.TargetFileID, v.Required, is.UUIDv4),
	)
}
