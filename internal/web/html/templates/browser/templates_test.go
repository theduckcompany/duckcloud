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
	}{
		{
			Name: "modal_create_dir",
			Template: &CreateDirTemplate{
				DirPath: "/foo/bar",
				SpaceID: uuid.UUID("some-space-id"),
				Error:   nil,
			},
		},
		{
			Name: "modal_create_dir with error",
			Template: &CreateDirTemplate{
				DirPath: "/foo/bar",
				SpaceID: uuid.UUID("some-space-id"),
				Error:   ptr.To("Some-error"),
			},
		},
		{
			Name: "modal_rename ",
			Template: &RenameTemplate{
				Error:               nil,
				Target:              &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo"},
				FieldValue:          "New Dir",
				FieldValueSelection: 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)
			r.Header.Add("HX-Boosted", "true")

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
