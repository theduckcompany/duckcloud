package dav

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dav/webdav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	davHandler *webdav.Handler
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, fs dfs.Service, folders folders.Service, davSessions davsessions.Service) *HTTPHandler {
	return &HTTPHandler{
		davHandler: &webdav.Handler{
			Prefix:     "/dav",
			FileSystem: fs,
			Folders:    folders,
			Sessions:   davSessions,
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, err error) {
				if err != nil {
					logger.LogEntrySetError(r, err)
				}
			},
		},
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.StripSlashed, mids.Logger)
	}

	r.Handle("/dav", h.davHandler)
	r.Handle("/dav/*", h.davHandler)
}

func (h *HTTPHandler) String() string {
	return "dav"
}
