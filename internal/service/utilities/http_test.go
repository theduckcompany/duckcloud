package utilities

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func Test_Robot_txt(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	srv := chi.NewRouter()
	NewHTTPHandler().Register(srv, nil)
	srv.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, "User-agent: *\nDisallow: /", w.Body.String())
}
