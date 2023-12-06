package browser

import (
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type CreateDirTemplate struct {
	DirPath string
	SpaceID uuid.UUID
	Error   *string
}

func (t *CreateDirTemplate) Template() string { return "browser/modal_create_dir.tmpl" }

type RenameTemplate struct {
	Error               *string
	Target              *dfs.PathCmd
	FieldValue          string
	FieldValueSelection int
}

func (t *RenameTemplate) Template() string { return "browser/modal_rename.tmpl" }

type RowsTemplate struct {
	Inodes []dfs.INode
	Folder *dfs.PathCmd
}

func (t *RowsTemplate) Template() string { return "browser/rows.tmpl" }

type BreadCrumbTemplate struct {
	Elements []BreadCrumbElement
}

func (t *BreadCrumbTemplate) Template() string { return "browser/breadcrumb.tmpl" }

type BreadCrumbElement struct {
	Name    string
	Href    string
	Current bool
}
