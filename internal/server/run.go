package server

import (
	"context"
	"database/sql"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"go.uber.org/fx"
)

func Run(ctx context.Context, db *sql.DB, fs afero.Fs) {
	// Start server with the HTTP server.
	app := start(ctx, db, fs, fx.Invoke(func(*router.API) {}))

	app.Run()
}
