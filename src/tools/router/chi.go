package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Registerer interface {
	Register(r chi.Router, mids Middlewares)
	String() string
}

// NewChiRouter return a new mux.Router with the generic middlewares.
//
// Those middlewares are:
// - Recoverer: Allow to catch the panics and recover.
// - CleanPath: Allow to remove the double slashs.
// - requestID: Generate an unique id for each request.
// - Timeout:   A context timeout after 60s.
func NewChiRouter(routes []Registerer, mids Middlewares) *chi.Mux {
	r := chi.NewMux()
	r.Use(
		middleware.Recoverer,
		middleware.CleanPath,
		middleware.RequestID,
		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		middleware.Timeout(60*time.Second),
	)

	for _, route := range routes {
		route.Register(r, mids)
	}

	return r
}
