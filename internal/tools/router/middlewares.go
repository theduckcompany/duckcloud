package router

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
)

type Middleware func(next http.Handler) http.Handler

type Middlewares struct {
	StripSlashed Middleware
	Logger       Middleware
	OnlyJSON     Middleware
	RealIP       Middleware
	CORS         Middleware
}

func InitMiddlewares(tools tools.Tools, configSvc config.Service) *Middlewares {
	return &Middlewares{
		StripSlashed: middleware.StripSlashes,
		Logger:       logger.NewRouterLogger(tools.Logger()),
		OnlyJSON:     middleware.AllowContentType("application/json"),
		RealIP:       middleware.RealIP,
		CORS: cors.Handler(cors.Options{
			AllowOriginFunc: func(r *http.Request, origin string) bool {
				hosts, err := configSvc.GetTrustedHosts(r.Context())
				if err != nil {
					return false
				}

				url, err := url.ParseRequestURI(origin)
				if err != nil {
					log.Printf("failed to parse the request uri: %s", err)
					return false
				}

				host, _, _ := net.SplitHostPort(url.Host)
				if host == "" {
					host = url.Host
				}

				return slices.Contains[[]string, string](hosts, host)
			},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	}
}
