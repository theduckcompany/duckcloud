package server

import (
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"go.uber.org/fx"
)

func Run(cfg *Config) {
	// Start server with the HTTP server.
	app := start(cfg, fx.Invoke(func(*router.API) {}))

	app.Run()
}
