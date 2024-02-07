package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
)

//go:embed public
var staticsFS embed.FS

type Config struct {
	HotReload bool `json:"hotReload"`
}

type HTTPHandler struct {
	cfg       Config
	assetFS   http.FileSystem
	startDate time.Time
}

func NewHTTPHandler(cfg Config) *HTTPHandler {
	var assetFS http.FileSystem

	switch cfg.HotReload {
	case true:
		assetFS = http.FS(os.DirFS("./assets/public"))
	case false:
		memFS, _ := fs.Sub(staticsFS, "public")
		assetFS = http.FS(memFS)
	}

	return &HTTPHandler{cfg, assetFS, time.Now()}
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r chi.Router, _ *router.Middlewares) {
	r.Get("/assets/*", h.handleAsset)
}

func (h *HTTPHandler) handleAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets")
	_, fileName := path.Split(assetPath)

	f, err := h.assetFS.Open(assetPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var lastModified time.Time
	if h.cfg.HotReload {
		// For the cache validation in dev mod
		w.Header().Add("Cache-Control", "no-cache")
		fileInfo, err := f.Stat()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		lastModified = fileInfo.ModTime()
	} else {
		// Expires header is for HTTP/1.1 and Cache-Controle is for the newer HTTP versions.
		w.Header().Add("Expires", time.Now().Add(365*24*time.Hour).UTC().Format(http.TimeFormat))
		w.Header().Add("Cache-Control", "max-age=31536000")
		lastModified = h.startDate
	}

	http.ServeContent(w, r, fileName, lastModified, f)
}
