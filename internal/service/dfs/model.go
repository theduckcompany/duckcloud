package dfs

import (
	io "io"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/users"
)

type INode = inodes.INode

var (
	ExampleAliceRoot  = inodes.ExampleAliceRoot
	ExampleAliceDir   = inodes.ExampleAliceDir
	ExampleAliceFile  = inodes.ExampleAliceFile
	ExampleAliceFile2 = inodes.ExampleAliceFile2
	ExampleBobRoot    = inodes.ExampleBobRoot
)

type UploadCmd struct {
	FilePath   string
	Content    io.Reader
	UploadedBy *users.User
}

func (t UploadCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.FilePath, v.Required, v.Length(1, 255)),
		v.Field(&t.Content, v.Required),
		v.Field(&t.UploadedBy, v.Required, v.NotNil),
	)
}

type CreateDirCmd struct {
	FilePath  string
	CreatedBy *users.User
}

func (t CreateDirCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.FilePath, v.Required, v.Length(1, 255)),
		v.Field(&t.CreatedBy, v.Required, v.NotNil),
	)
}

type MoveCmd struct {
	SrcPath string
	NewPath string
	MovedBy *users.User
}

func (t MoveCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.SrcPath, v.Required, v.Length(1, 255)),
		v.Field(&t.NewPath, v.Required, v.Length(1, 255)),
		v.Field(&t.MovedBy, v.Required, v.NotNil),
	)
}
