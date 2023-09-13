package router

import (
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/router/internal"
)

type Middleware func(next http.Handler) http.Handler

type Middlewares struct {
	StripSlashed Middleware
	Logger       Middleware
	OnlyJSON     Middleware
	RealIP       Middleware
	CORS         Middleware
}

func InitMiddlewares(tools tools.Tools) *Middlewares {
	return &Middlewares{
		StripSlashed: middleware.StripSlashes,
		Logger:       internal.NewStructuredLogger(tools.Logger()),
		OnlyJSON:     middleware.AllowContentType("application/json"),
		RealIP:       middleware.RealIP,
		CORS: cors.Handler(cors.Options{
			AllowOriginFunc: func(r *http.Request, origin string) bool {
				cleanPath := path.Clean(r.URL.Path)
				// Allows all the routes excepts the auth one to be accessed by
				// some other domain name
				if strings.Contains(cleanPath, "/login") ||
					strings.Contains(cleanPath, "/forgot") ||
					strings.Contains(cleanPath, "/consent") {
					return false
				}
				return true
			},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	}
}
