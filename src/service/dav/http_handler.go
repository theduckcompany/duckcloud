package dav

import (
	"github.com/Peltoche/neurone/src/service/dav/internal"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"golang.org/x/net/webdav"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	log *slog.Logger
	fs  Service
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools, fs Service) *HTTPHandler {
	return &HTTPHandler{log: tools.Logger(), fs: fs}
}

func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	dav := r.With(mids.StripSlashed, mids.Logger)

	dav.Handle("/dav/*", &webdav.Handler{
		Prefix:     "/dav/",
		FileSystem: h.fs,
		LockSystem: webdav.NewMemLS(),
		Logger:     internal.NewLogger(h.log),
	})
}

func (h *HTTPHandler) String() string {
	return "dav"
}
