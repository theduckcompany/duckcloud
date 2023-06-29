package dav

import (
	"io"
	"net/http"

	"github.com/Peltoche/neurone/src/service/dav/internal"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"golang.org/x/net/webdav"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	log *slog.Logger
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(tools tools.Tools) *HTTPHandler {
	return &HTTPHandler{log: tools.Logger()}
}

func (h *HTTPHandler) Register(r *chi.Mux) {
	dav := webdav.Handler{
		Prefix:     "/dav/",
		FileSystem: webdav.Dir("./testdata"),
		LockSystem: webdav.NewMemLS(),
		Logger:     internal.NewLogger(h.log),
	}

	r.Handle("/dav/", &dav)
}

func (h *HTTPHandler) String() string {
	return "dav"
}

func (h *HTTPHandler) echoHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.ErrorCtx(r.Context(), "Failed to handle request:", err)
	}
}
