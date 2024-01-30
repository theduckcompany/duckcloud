package browser

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func Test_Templates(t *testing.T) {
	renderer := html.NewRenderer(html.Config{
		PrettyRender: false,
		HotReload:    false,
	})

	tests := []struct {
		Name     string
		Template html.Templater
		Layout   bool
	}{
		{
			Name:   "modal_create_dir",
			Layout: false,
			Template: &CreateDirTemplate{
				DirPath: "/foo/bar",
				SpaceID: uuid.UUID("some-space-id"),
				Error:   nil,
			},
		},
		{
			Name:   "modal_create_dir with error",
			Layout: false,
			Template: &CreateDirTemplate{
				DirPath: "/foo/bar",
				SpaceID: uuid.UUID("some-space-id"),
				Error:   ptr.To("Some-error"),
			},
		},
		{
			Name:   "modal_rename",
			Layout: false,
			Template: &RenameTemplate{
				Error:               nil,
				Target:              dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				FieldValue:          "New Dir",
				FieldValueSelection: 0,
			},
		},
		{
			Name:   "rows",
			Layout: false,
			Template: &RowsTemplate{
				Folder:        dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				Inodes:        []dfs.INode{dfs.ExampleAliceFile, dfs.ExampleAliceFile2},
				ContentTarget: "#content",
			},
		},
		{
			Name:   "breadcrumb",
			Layout: false,
			Template: &BreadCrumbTemplate{
				Elements: []BreadCrumbElement{
					{Name: "My Files", Href: "https://localhost/", Current: false},
					{Name: "foo", Href: "https://localhost/foo", Current: true},
				},
				Target: "#content",
			},
		},
		{
			Name:   "content",
			Layout: true,
			Template: &ContentTemplate{
				Folder:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				Inodes:       []dfs.INode{dfs.ExampleAliceFile, dfs.ExampleAliceFile2},
				CurrentSpace: &spaces.ExampleAlicePersonalSpace,
				AllSpaces:    []spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleAliceBobSharedSpace},
			},
		},
		{
			Name:   "move modal",
			Layout: false,
			Template: &MoveTemplate{
				SrcPath:  dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				SrcInode: &dfs.ExampleAliceDir,
				DstPath:  dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar"),
				FolderContent: map[dfs.PathCmd]dfs.INode{
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file1.jpg"): dfs.ExampleAliceFile,
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file2.jpg"): dfs.ExampleAliceFile2,
				},
				PageSize: 10,
			},
		},
		{
			Name:   "move rows",
			Layout: false,
			Template: &MoveRowsTemplate{
				SrcPath: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				DstPath: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar"),
				FolderContent: map[dfs.PathCmd]dfs.INode{
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file1.jpg"): dfs.ExampleAliceFile,
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file2.jpg"): dfs.ExampleAliceFile2,
				},
				PageSize: 10,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)

			if !test.Layout {
				r.Header.Add("HX-Boosted", "true")
			}

			renderer.WriteHTMLTemplate(w, r, http.StatusOK, test.Template)

			if !assert.Equal(t, http.StatusOK, w.Code) {
				res := w.Result()
				res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				t.Log(string(body))
			}
		})
	}
}

func TestMoveTemplateBreadcrumb(t *testing.T) {
	tests := []struct {
		Name     string
		Path     *dfs.PathCmd
		Expected []BreadCrumbElement
	}{
		{
			Name: "Simple",
			Path: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar"),
			Expected: []BreadCrumbElement{
				{
					Name:    spaces.ExampleAlicePersonalSpace.Name(),
					Href:    "/browser/move?dstPath=%2F&spaceID=e97b60f7-add2-43e1-a9bd-e2dac9ce69ec&srcPath=%2Ffoo",
					Current: false,
				},
				{
					Name:    "bar",
					Href:    "/browser/move?dstPath=%2Fbar&spaceID=e97b60f7-add2-43e1-a9bd-e2dac9ce69ec&srcPath=%2Ffoo",
					Current: true,
				},
			},
		},
		{
			Name: "Space root",
			Path: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/"),
			Expected: []BreadCrumbElement{
				{
					Name:    spaces.ExampleAlicePersonalSpace.Name(),
					Href:    "/browser/move?dstPath=%2F&spaceID=e97b60f7-add2-43e1-a9bd-e2dac9ce69ec&srcPath=%2Ffoo",
					Current: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			moveTemplate := MoveTemplate{
				SrcPath:  dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
				SrcInode: &dfs.ExampleAliceDir,
				DstPath:  test.Path, // The breadcrumb is created based on the destination path.
				FolderContent: map[dfs.PathCmd]dfs.INode{
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file1.jpg"): dfs.ExampleAliceFile,
					*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file2.jpg"): dfs.ExampleAliceFile2,
				},
				PageSize: 10,
			}

			res := moveTemplate.Breadcrumb()

			assert.Equal(t, test.Expected, res.Elements)
		})
	}
}
