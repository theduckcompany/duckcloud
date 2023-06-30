package assets

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed assets
var staticsFS embed.FS

type Config struct {
	HotReload bool `mapstructure:"hotReload"`
}

type HTTPHandler struct {
	cfg Config
}

func NewHTTPHandler(cfg Config) *HTTPHandler {
	return &HTTPHandler{cfg}
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r *chi.Mux) {
	var server http.Handler

	switch h.cfg.HotReload {
	case true:
		fs := http.Dir("./src/service/assets/assets")
		server = http.StripPrefix("/assets", http.FileServer(fs))
	case false:
		fs := http.FS(staticsFS)
		server = http.FileServer(fs)
	}

	r.Get("/assets/*", http.HandlerFunc(server.ServeHTTP))

}

func (h *HTTPHandler) String() string {
	return "assets"
}
