package dav

import (
	"io"
	"net/http"

	"github.com/Peltoche/neurone/pkg/service/dav/internal"
	"github.com/Peltoche/neurone/pkg/tools/logger"
	"golang.org/x/net/webdav"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	log *logger.Logger
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(log *logger.Logger) *HTTPHandler {
	return &HTTPHandler{log}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	dav := webdav.Handler{
		Prefix:     "/dav/",
		FileSystem: webdav.Dir("./testdata"),
		LockSystem: webdav.NewMemLS(),
		Logger:     internal.NewLogger(h.log),
	}

	mux.Handle("/dav/", &dav)
}

func (h *HTTPHandler) String() string {
	return "dav"
}

func (h *HTTPHandler) echoHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.ErrorCtx(r.Context(), "Failed to handle request:", err)
	}
}
