package server

import (
	"context"
	"database/sql"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"go.uber.org/fx"
)

func Run(ctx context.Context, db *sql.DB, fs afero.Fs, folderPath string) {
	// Start server with the HTTP server.
	app := start(ctx, db, fs, folderPath, fx.Invoke(func(*router.API, runner.Service) {}))

	app.Run()
}
