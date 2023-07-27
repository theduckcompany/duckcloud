package router

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/router/internal"
	"github.com/go-chi/chi/v5/middleware"
)

type Middleware func(next http.Handler) http.Handler

type Middlewares struct {
	StripSlashed Middleware
	Logger       Middleware
	OnlyJSON     Middleware
	RealIP       Middleware
}

func InitMiddlewares(tools tools.Tools) Middlewares {
	return Middlewares{
		StripSlashed: middleware.StripSlashes,
		Logger:       internal.NewStructuredLogger(tools.Logger()),
		OnlyJSON:     middleware.AllowContentType("application/json"),
		RealIP:       middleware.RealIP,
	}
}
