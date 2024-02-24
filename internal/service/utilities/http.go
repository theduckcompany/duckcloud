package utilities

import (
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
)

type HTTPHandler struct{}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (t *HTTPHandler) Register(r chi.Router, _ *router.Middlewares) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
