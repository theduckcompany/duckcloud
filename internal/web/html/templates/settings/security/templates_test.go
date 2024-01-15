package security

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
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
			Name:   "ContentTemplate",
			Layout: true,
			Template: &ContentTemplate{
				IsAdmin:        false,
				CurrentSession: &websessions.AliceWebSessionExample,
				WebSessions:    []websessions.Session{websessions.AliceWebSessionExample},
				Devices:        []davsessions.DavSession{davsessions.ExampleAliceSession},
				Spaces: map[uuid.UUID]spaces.Space{
					spaces.ExampleAlicePersonalSpace.ID(): spaces.ExampleAlicePersonalSpace,
				},
			},
		},
		{
			Name:   "PasswordFormTemplate",
			Layout: false,
			Template: &PasswordFormTemplate{
				Error: "some-error",
			},
		},
		{
			Name:   "WebdavFormTemplate",
			Layout: false,
			Template: &WebdavFormTemplate{
				Error:  nil,
				Spaces: []spaces.Space{spaces.ExampleBobPersonalSpace},
			},
		},
		{
			Name:   "WebdavResultTemplate",
			Layout: false,
			Template: &WebdavResultTemplate{
				Secret:     "some-secret",
				NewSession: &davsessions.ExampleAliceSession,
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
