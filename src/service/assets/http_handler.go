package assets

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed public
var staticsFS embed.FS

type HTTPHandler struct {
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r *chi.Mux) {
	r.Get("/assets/", http.FileServer(http.FS(staticsFS)).ServeHTTP)
}

func (h *HTTPHandler) String() string {
	return "assets"
}
