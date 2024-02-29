package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func Test_Templates(t *testing.T) {
	renderer := html.NewRenderer(html.Config{
		PrettyRender: false,
		HotReload:    false,
	})

	tests := []struct {
		Template html.Templater
		Name     string
		Layout   bool
	}{
		{
			Name:   "LoginPageTmpl",
			Layout: true,
			Template: &LoginPageTmpl{
				UsernameContent: "some-user-input",
				UsernameError:   "some-error-msg",
				PasswordError:   "",
			},
		},
		{
			Name:   "ErrorPageTmpl",
			Layout: true,
			Template: &ErrorPageTmpl{
				ErrorMsg:  "some-error",
				RequestID: "some-request-id",
			},
		},
		{
			Name:   "ConsentPageTmpl",
			Layout: true,
			Template: &ConsentPageTmpl{
				Username:   "Alice",
				Redirect:   "/foo/bar",
				ClientName: "some-name",
				Scopes:     []string{"a.b", "c.d"},
			},
		},
		{
			Name:   "AskMasterPassword",
			Layout: true,
			Template: &AskMasterPasswordPageTmpl{
				ErrorMsg: "some message",
			},
		},
		{
			Name:   "RegisterMasterPassword",
			Layout: true,
			Template: &RegisterMasterPasswordPageTmpl{
				ErrorMsg: "some message",
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
