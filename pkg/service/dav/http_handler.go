package dav

import (
	"io"
	"net/http"

	"github.com/Peltoche/neurone/pkg/tools/logger"
)

// HTTPHandler serve files via the Webdav protocol over http.
type HTTPHandler struct {
	logger *logger.Logger
}

// NewHTTPHandler builds a new EchoHandler.
func NewHTTPHandler(logger *logger.Logger) *HTTPHandler {
	return &HTTPHandler{logger}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/dav", h.echoHandler)
}

func (h *HTTPHandler) String() string {
	return "dav"
}

func (h *HTTPHandler) echoHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to handle request:", err)
	}
}
