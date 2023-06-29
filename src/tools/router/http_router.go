package router

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// MuxHandler is an http.Handler that knows the mux pattern
// under which it will be registered.
type MuxHandler interface {
	Register(*http.ServeMux)

	// Strings reports the handler name
	String() string
}

func NewServer(lc fx.Lifecycle, handler *chi.Mux, tools tools.Tools) *http.Server {
	logger := tools.Logger()

	srv := &http.Server{Addr: ":8080", Handler: handler}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			logger.Info(fmt.Sprintf("Starting HTTP server at %s", srv.Addr))

			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
