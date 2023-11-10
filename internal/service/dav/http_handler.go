package dav

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dav/webdav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	webdavHandler *webdav.Handler
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, fs dfs.Service, files files.Service, folders folders.Service, davSessions davsessions.Service) *HTTPHandler {
	return &HTTPHandler{
		webdavHandler: &webdav.Handler{
			Prefix:     "/webdav",
			FileSystem: fs,
			Folders:    folders,
			Files:      files,
			Sessions:   davSessions,
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

	r.HandleFunc("/webdav", h.handleWebdavCollections)
	r.Handle("/webdav/*", h.webdavHandler)
}

func (h *HTTPHandler) String() string {
	return "dav"
}

func (h *HTTPHandler) handleWebdavCollections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		h.webdavHandler.ServeHTTP(w, r)
	default:
	}
}
