package router

import (
	"fmt"
	"time"

	"github.com/Peltoche/neurone/src/tools/router/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

type Registerer interface {
	Register(r *chi.Mux)
	String() string
}

// NewChiRouter return a new mux.Router with the basic setup.
func NewChiRouter(routes []Registerer, log *slog.Logger) *chi.Mux {

	r := chi.NewMux()
	r.Use(
		internal.NewStructuredLogger(log),
		middleware.Recoverer,
		middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"),
		middleware.StripSlashes,
		middleware.CleanPath,
		middleware.Compress(5, "application/json"),
		middleware.RequestID,
	)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	for _, route := range routes {
		route.Register(r)
		log.Info(fmt.Sprintf("Register %q", route.String()))

	}

	return r
}
