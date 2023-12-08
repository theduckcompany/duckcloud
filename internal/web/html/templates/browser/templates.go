package browser

import (
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type LayoutTemplate struct {
	CurrentSpace *spaces.Space
	Spaces       []spaces.Space
}

func (t *LayoutTemplate) Template() string { return "browser/layout.tmpl" }

type ContentTemplate struct {
	Folder       *dfs.PathCmd
	Inodes       []dfs.INode
	CurrentSpace *spaces.Space
	AllSpaces    []spaces.Space
}

func (t *ContentTemplate) Template() string { return "browser/content.tmpl" }

func (t *ContentTemplate) Breadcrumb() *BreadCrumbTemplate {
	return &BreadCrumbTemplate{Path: t.Folder}
}

func (t *ContentTemplate) Rows() *RowsTemplate {
	return &RowsTemplate{
		Folder: t.Folder,
		Inodes: t.Inodes,
	}
}

func (t *ContentTemplate) Layout() *LayoutTemplate {
	return &LayoutTemplate{
		CurrentSpace: t.CurrentSpace,
		Spaces:       t.AllSpaces,
	}
}

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
	Path *dfs.PathCmd
}

func (t *BreadCrumbTemplate) Template() string { return "browser/breadcrumb.tmpl" }

func (t *BreadCrumbTemplate) Elements() []BreadCrumbElement {
	basePath := path.Join("/browser/", string(t.Path.Space.ID()))

	elements := []BreadCrumbElement{{
		Name:    t.Path.Space.Name(),
		Href:    basePath,
		Current: false,
	}}

	fullPath := strings.TrimPrefix(t.Path.Path, "/")

	if fullPath == "" {
		elements[0].Current = true
		return elements
	}

	for _, elem := range strings.Split(fullPath, "/") {
		basePath = path.Join(basePath, elem)

		elements = append(elements, BreadCrumbElement{
			Name:    elem,
			Href:    basePath,
			Current: false,
		})
	}

	elements[len(elements)-1].Current = true

	return elements
}

type BreadCrumbElement struct {
	Name    string
	Href    string
	Current bool
}
