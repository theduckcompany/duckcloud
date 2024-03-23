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
	CurrentSpace  *spaces.Space
	ContentTarget string
	Inodes        []dfs.INode
	AllSpaces     []spaces.Space
}

func (t *ContentTemplate) Template() string { return "browser/page" }

func (t *ContentTemplate) Breadcrumb() *BreadCrumbTemplate {
	basePath := path.Join("/browser/", string(t.Folder.Space().ID()))

	elements := []BreadCrumbElement{{
		Name: t.Folder.Space().Name(),
		Href: basePath,
	}}

	fullPath := strings.TrimPrefix(t.Folder.Path(), "/")

	if fullPath == "" {
		return &BreadCrumbTemplate{
			Parents:    []BreadCrumbElement{},
			CurrentDir: elements[0],
			Target:     "body",
		}
	}

	for _, elem := range strings.Split(fullPath, "/") {
		basePath = path.Join(basePath, elem)

		elements = append(elements, BreadCrumbElement{
			Name: elem,
			Href: basePath,
		})
	}

	return &BreadCrumbTemplate{
		Parents:    elements[:len(elements)-1],
		CurrentDir: elements[len(elements)-1],
		Target:     "body",
	}
}

func (t *ContentTemplate) Rows() *RowsTemplate {
	return &RowsTemplate{
		Folder:        t.Folder,
		Inodes:        t.Inodes,
		ContentTarget: t.ContentTarget,
	}
}

type CreateDirTemplate struct {
	Error   *string
	DirPath string
	SpaceID uuid.UUID
}

func (t *CreateDirTemplate) Template() string { return "browser/modal_create_dir" }

type RenameTemplate struct {
	Error               *string
	Target              *dfs.PathCmd
	FieldValue          string
	FieldValueSelection int
}

func (t *RenameTemplate) Template() string { return "browser/modal_rename" }

type RowsTemplate struct {
	Folder        *dfs.PathCmd
	ContentTarget string
	Inodes        []dfs.INode
}

func (t *RowsTemplate) Template() string { return "browser/rows" }

type BreadCrumbTemplate struct {
	CurrentDir BreadCrumbElement
	Target     string
	Parents    []BreadCrumbElement
}

func (t *BreadCrumbTemplate) Template() string { return "browser/breadcrumb" }

type BreadCrumbElement struct {
	Name string
	Href string
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
		Name: t.SrcPath.Space().Name(),
		Href: basePath.String(),
	}}

	if t.DstPath.Path() == "/" {
		return &BreadCrumbTemplate{
			Parents:    []BreadCrumbElement{},
			CurrentDir: elements[0],
			Target:     "#modal-content",
		}
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
			Name: elem,
			Href: basePath.String(),
		})
	}

	return &BreadCrumbTemplate{
		Parents:    elements[:len(elements)-1],
		CurrentDir: elements[len(elements)-1],
		Target:     "#modal-content",
	}
}

func (t *MoveTemplate) Template() string { return "browser/modal_move" }

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

func (t *MoveRowsTemplate) Template() string { return "browser/modal_move_rows" }

type MediaViewerModal struct {
	Path     *dfs.PathCmd
	FileName string
	Folder   string
}

func (t *MediaViewerModal) Template() string { return "browser/modal_media_viewer" }
