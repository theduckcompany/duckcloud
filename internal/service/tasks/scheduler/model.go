package scheduler

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FileUploadArgs struct {
	FolderID   uuid.UUID `json:"folder-id"`
	Directory  uuid.UUID `json:"directory"`
	FileID     uuid.UUID `json:"file-id"`
	FileName   string    `json:"file-name"`
	UploadedAt time.Time `json:"uploaded-at"`
}

func (a FileUploadArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.FolderID, v.Required, is.UUIDv4),
		v.Field(&a.Directory, v.Required, is.UUIDv4),
		v.Field(&a.FileID, v.Required, is.UUIDv4),
		v.Field(&a.FileName, v.Required, v.Length(1, 1024)),
		v.Field(&a.UploadedAt, v.Required),
	)
}

type FileMoveArgs struct {
	FolderID   uuid.UUID `json:"folder-id"`
	INodeID    uuid.UUID `json:"inode-id"`
	TargetPath string    `json:"target-path"`
	MovedAt    time.Time `json:"moved-at"`
}

func (a FileMoveArgs) Validate() error {
	return v.ValidateStruct(&a,
		v.Field(&a.FolderID, v.Required, is.UUIDv4),
		v.Field(&a.INodeID, v.Required, is.UUIDv4),
		v.Field(&a.TargetPath, v.Required),
	)
}

type FSGCArgs struct{}

func (a *FSGCArgs) Validate() error {
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
