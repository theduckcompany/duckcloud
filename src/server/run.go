package server

import (
	"net/http"

	"go.uber.org/fx"
)

func Run(cfg *Config) {
	// Start server with the HTTP server.
	app := start(cfg, fx.Invoke(func(*http.Server) {}))

	app.Run()
}
