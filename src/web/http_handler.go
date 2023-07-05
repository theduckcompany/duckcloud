package web

import (
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	response response.Writer
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	_ = r.With(mids.StripSlashed)
}

func (h *HTTPHandler) String() string {
	return "web"
}
