package browser

import (
	"net/url"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type ContentTemplate struct {
	Folder        *dfs.PathCmd
	Inodes        []dfs.INode
	CurrentSpace  *spaces.Space
	AllSpaces     []spaces.Space
	ContentTarget string
}

func (t *ContentTemplate) Template() string { return "browser/content.tmpl" }

func (t *ContentTemplate) Breadcrumb() *BreadCrumbTemplate {
	basePath := path.Join("/browser/", string(t.Folder.Space().ID()))

	elements := []BreadCrumbElement{{
		Name:    t.Folder.Space().Name(),
		Href:    basePath,
		Current: false,
	}}

	fullPath := strings.TrimPrefix(t.Folder.Path(), "/")

	if fullPath == "" {
		elements[0].Current = true
		return &BreadCrumbTemplate{Elements: elements, Target: "#content"}
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

	return &BreadCrumbTemplate{Elements: elements, Target: "#content"}
}

func (t *ContentTemplate) Rows() *RowsTemplate {
	return &RowsTemplate{
		Folder:        t.Folder,
		Inodes:        t.Inodes,
		ContentTarget: t.ContentTarget,
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
	Inodes        []dfs.INode
	Folder        *dfs.PathCmd
	ContentTarget string
}

func (t *RowsTemplate) Template() string { return "browser/rows.tmpl" }

type BreadCrumbTemplate struct {
	Elements []BreadCrumbElement
	Target   string
}

func (t *BreadCrumbTemplate) Template() string { return "browser/breadcrumb.tmpl" }

type BreadCrumbElement struct {
	Name    string
	Href    string
	Current bool
}

type MoveTemplate struct {
	SrcPath       *dfs.PathCmd
	SrcInode      *dfs.INode
	DstPath       *dfs.PathCmd
	FolderContent map[dfs.PathCmd]dfs.INode
	PageSize      int
}

func (t *MoveTemplate) Breadcrumb() *BreadCrumbTemplate {
	vals := url.Values{
		"srcPath": []string{t.SrcPath.Path()},
		"dstPath": []string{"/"},
		"spaceID": []string{string(t.SrcPath.Space().ID())},
	}

	basePath := url.URL{Path: "/browser/move", RawQuery: vals.Encode()}

	elements := []BreadCrumbElement{{
		Name:    t.SrcPath.Space().Name(),
		Href:    basePath.String(),
		Current: false,
	}}

	if t.DstPath.Path() == "/" {
		elements[0].Current = true
		return &BreadCrumbTemplate{Elements: elements, Target: "#modal-content"}
	}

	dstPath := "/"
	for _, elem := range strings.Split(strings.TrimPrefix(t.DstPath.Path(), "/"), "/") {
		dstPath = path.Join(dstPath, elem)

		vals := url.Values{
			"srcPath": []string{t.SrcPath.Path()},
			"dstPath": []string{dstPath},
			"spaceID": []string{string(t.SrcPath.Space().ID())},
		}

		basePath := url.URL{Path: "/browser/move", RawQuery: vals.Encode()}

		elements = append(elements, BreadCrumbElement{
			Name:    elem,
			Href:    basePath.String(),
			Current: false,
		})
	}

	elements[len(elements)-1].Current = true

	return &BreadCrumbTemplate{Elements: elements, Target: "#modal-content"}
}

func (t *MoveTemplate) Template() string { return "browser/modal_move.tmpl" }

func (t *MoveTemplate) MoveRows() *MoveRowsTemplate {
	return &MoveRowsTemplate{
		SrcPath:       t.SrcPath,
		DstPath:       t.DstPath,
		FolderContent: t.FolderContent,
		PageSize:      t.PageSize,
	}
}

type MoveRowsTemplate struct {
	SrcPath       *dfs.PathCmd
	DstPath       *dfs.PathCmd
	FolderContent map[dfs.PathCmd]dfs.INode
	PageSize      int
}

func (t *MoveRowsTemplate) Template() string { return "browser/modal_move_rows.tmpl" }
