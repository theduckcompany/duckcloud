package dav

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dav/internal"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"golang.org/x/net/webdav"
)

type davKeyCtx string

var sessionKeyCtx davKeyCtx = "user"

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	davSessions davsessions.Service
	davHandler  *webdav.Handler
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, fs dfs.Service, folders folders.Service, davSessions davsessions.Service) *HTTPHandler {
	return &HTTPHandler{
		davSessions: davSessions,
		davHandler: &webdav.Handler{
			Prefix:     "/dav",
			FileSystem: &davFS{folders, fs},
			LockSystem: webdav.NewMemLS(),
			Logger:     internal.NewLogger(tools.Logger()),
		},
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.StripSlashed, mids.Logger)
	}

	r.HandleFunc("/dav", h.handle)
	r.HandleFunc("/dav/*", h.handle)
}

func (h *HTTPHandler) String() string {
	return "dav"
}

func (h *HTTPHandler) handle(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", `Basic realm="fs"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := h.davSessions.Authenticate(r.Context(), username, password)
	if errors.Is(err, davsessions.ErrInvalidCredentials) {
		w.Header().Add("WWW-Authenticate", `Basic realm="fs"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), sessionKeyCtx, session)

	h.davHandler.ServeHTTP(w, r.WithContext(ctx))
}
