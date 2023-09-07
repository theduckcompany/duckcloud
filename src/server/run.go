package server

import (
	"database/sql"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"go.uber.org/fx"
)

func Run(cfg *Config, db *sql.DB, fs afero.Fs) {
	// Start server with the HTTP server.
	app := start(cfg, db, fs, fx.Invoke(func(*router.API) {}))

	app.Run()
}
