package httprouter

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/Peltoche/neurone/src/tools/logger"
	"go.uber.org/fx"
)

// MuxHandler is an http.Handler that knows the mux pattern
// under which it will be registered.
type MuxHandler interface {
	Register(*http.ServeMux)

	// Strings reports the handler name
	String() string
}

func NewServer(lc fx.Lifecycle, mux *http.ServeMux, log *logger.Logger) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Starting HTTP server at %s", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
