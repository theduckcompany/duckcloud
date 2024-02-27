package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func Test_RedirectionsSettings(t *testing.T) {
	t.Run("redirectDefaultSettings success", func(t *testing.T) {
		handler := NewRedirections()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusMovedPermanently, res.StatusCode)
		assert.Equal(t, "/settings/security", res.Header.Get("Location"))
	})
}
