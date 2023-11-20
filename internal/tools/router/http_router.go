package router

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/coreos/go-systemd/daemon"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"go.uber.org/fx"
)

type API struct{}

type Config struct {
	Addr      string
	TLS       bool
	CertFile  string
	KeyFile   string
	HostNames []string
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
	chi.RegisterMethod("MKCALENDAR")
	chi.RegisterMethod("MKCOL")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("OPTIONS")
	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("REPORT")
	chi.RegisterMethod("SEARCH")
	chi.RegisterMethod("UNCHECKOUT")
	chi.RegisterMethod("VERSION-CONTROL")
}

func NewServer(routes []Registerer, cfg Config, lc fx.Lifecycle, mids *Middlewares, tools tools.Tools, fs afero.Fs) (*API, error) {
	handler, err := createHandler(cfg, routes, mids)
	if err != nil {
		return nil, fmt.Errorf("failed to create the listener: %w", err)
	}

	httpLogger := slog.NewLogLogger(tools.Logger().Handler(), slog.LevelError)

	srv := &http.Server{
		Addr:     cfg.Addr,
		Handler:  handler,
		ErrorLog: httpLogger,
	}

	if cfg.TLS {
		cert, err := afero.ReadFile(fs, cfg.CertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load the TLS certification file: %w", err)
		}

		key, err := afero.ReadFile(fs, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load the TLS key file: %w", err)
		}

		certif, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to generate the X509 key pair: %w", err)
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certif},
		}
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ln, err := net.Listen("tcp", cfg.Addr)
			if err != nil {
				return err
			}

			tools.Logger().Info("start listening", slog.String("host", ln.Addr().String()), slog.Int("routes", len(handler.Routes())))
			if cfg.TLS {
				go srv.ServeTLS(ln, "", "")
			} else {
				go srv.Serve(ln)
			}

			for _, route := range handler.Routes() {
				tools.Logger().Debug("expose endpoint", slog.String("host", ln.Addr().String()), slog.String("route", route.Pattern))
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	daemon.SdNotify(false, daemon.SdNotifyReady)

	return &API{}, nil
}

func createHandler(cfg Config, routes []Registerer, mids *Middlewares) (chi.Router, error) {
	r := chi.NewMux()
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
	})

	if cfg.TLS {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
				w.Header().Set("Strict-Transport-Security", "max-age=15768000; preload")
			})
		})
	}

	r.Use(mids.CORS)
	r.Use(middleware.RequestID)

	for _, svc := range routes {
		svc.Register(r, mids)
	}

	return r, nil
}
