package dav

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/dav/internal"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"golang.org/x/net/webdav"
)

type davKeyCtx string

var userKeyCtx davKeyCtx = "user"

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	users      users.Service
	davHandler *webdav.Handler
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, inodes inodes.Service, files files.Service, users users.Service) *HTTPHandler {
	return &HTTPHandler{
		users: users,
		davHandler: &webdav.Handler{
			Prefix:     "/dav",
			FileSystem: &davFS{inodes, files},
			LockSystem: webdav.NewMemLS(),
			Logger:     internal.NewLogger(tools.Logger()),
		},
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	dav := r.With(mids.StripSlashed, mids.Logger)

	dav.HandleFunc("/dav", h.handle)
	dav.HandleFunc("/dav/*", h.handle)
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

	user, err := h.users.Authenticate(r.Context(), username, password)
	if errors.Is(err, users.ErrInvalidUsername) || errors.Is(err, users.ErrInvalidPassword) {
		w.Header().Add("WWW-Authenticate", `Basic realm="fs"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(r.Context(), userKeyCtx, user)

	h.davHandler.ServeHTTP(w, r.WithContext(ctx))
}
