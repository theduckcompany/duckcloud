package users

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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
			Name:   "ContentTemplate",
			Layout: true,
			Template: &ContentTemplate{
				IsAdmin: true,
				Current: &users.ExampleAlice,
				Users:   []users.User{users.ExampleAlice, users.ExampleBob},
				Error:   nil,
			},
		},
		{
			Name:   "RegistrationFormTemplate",
			Layout: false,
			Template: &RegistrationFormTemplate{
				Error: fmt.Errorf("some-error"),
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
