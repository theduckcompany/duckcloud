package router

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"go.uber.org/fx"
)

type API struct{}

type Config struct {
	Port          int      `json:"port"`
	TLS           bool     `json:"tls"`
	BindAddresses []string `json:"bindAddresses"`
	Services      []string `json:"services"`
}

type Registerer interface {
	Register(r chi.Router, mids *Middlewares)
	String() string
}

//nolint:gochecknoinits // This is the only way to ensure that we register the methods only once into the global router
func init() {
	chi.RegisterMethod("ACL")
	chi.RegisterMethod("CANCELUPLOAD")
	chi.RegisterMethod("CHECKIN")
	chi.RegisterMethod("CHECKOUT")
	chi.RegisterMethod("COPY")
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("MKCALENDAR")
	chi.RegisterMethod("MKCOL")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("OPTIONS")
	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("REPORT")
	chi.RegisterMethod("SEARCH")
	chi.RegisterMethod("UNCHECKOUT")
	chi.RegisterMethod("UNLOCK")
	chi.RegisterMethod("VERSION-CONTROL")
}

func NewServer(routes []Registerer, cfgs []Config, lc fx.Lifecycle, mids *Middlewares, tools tools.Tools) (*API, error) {
	for idx, cfg := range cfgs {
		handler, err := createHandler(cfg, routes, mids)
		if err != nil {
			return nil, fmt.Errorf("failed to create the listener n'%d: %w", idx, err)
		}

		for _, addr := range cfg.BindAddresses {
			hostPort := net.JoinHostPort(addr, strconv.Itoa(cfg.Port))

			srv := &http.Server{
				Addr:    hostPort,
				Handler: handler,
			}

			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					ln, err := net.Listen("tcp", hostPort)
					if err != nil {
						return err
					}

					go srv.Serve(ln)

					tools.Logger().Info("start listening", slog.String("host", ln.Addr().String()), slog.Int("routes", len(handler.Routes())))
					for _, route := range handler.Routes() {
						tools.Logger().Debug("expose endpoint", slog.String("host", ln.Addr().String()), slog.String("route", route.Pattern))
					}

					return nil
				},
				OnStop: func(ctx context.Context) error {
					return srv.Shutdown(ctx)
				},
			})
		}
	}

	return &API{}, nil
}

func createHandler(cfg Config, routes []Registerer, mids *Middlewares) (chi.Router, error) {
	r := chi.NewMux()
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
	})
	r.Use(middleware.RequestID)

	for _, svc := range cfg.Services {
		svcIdx := -1
		for idx, route := range routes {
			if route.String() == svc {
				svcIdx = idx
				break
			}
		}
		if svcIdx == -1 {
			return nil, fmt.Errorf("unknown service %q", svc)
		}

		routes[svcIdx].Register(r, mids)
	}

	return r, nil
}
