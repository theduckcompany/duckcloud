package dav

import (
	"context"
	"net/http"

	"github.com/Peltoche/neurone/src/service/dav/internal"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/webdav"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	davHandler *webdav.Handler
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, fs Service) *HTTPHandler {
	return &HTTPHandler{
		davHandler: &webdav.Handler{
			Prefix:     "/dav/",
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
			Logger:     internal.NewLogger(tools.Logger()),
		},
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	dav := r.With(mids.StripSlashed, mids.Logger)

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
	}

	ctx := context.WithValue(r.Context(), usernameKeyCtx, username)
	ctx = context.WithValue(ctx, passwordKeyCtx, password)

	h.davHandler.ServeHTTP(w, r.WithContext(ctx))
}
