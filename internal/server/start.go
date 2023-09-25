package server

import (
	"database/sql"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/jobs"
	"github.com/theduckcompany/duckcloud/internal/service/dav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/debug"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauth2"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/web"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// AsRoute annotates the given constructor to state that
// it provides a route to the "routes" group.
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(router.Registerer)),
		fx.ResultTags(`group:"routes"`),
	)
}

func start(cfg *Config, db *sql.DB, fs afero.Fs, invoke fx.Option) *fx.App {
	app := fx.New(
		fx.WithLogger(func(tools tools.Tools) fxevent.Logger { return logger.NewFxLogger(tools.Logger()) }),
		fx.Provide(
			func() Config { return *cfg },
			func() afero.Fs { return fs },
			func() *sql.DB { return db },

			// Tools
			fx.Annotate(tools.NewToolbox, fx.As(new(tools.Tools))),

			// Services
			fx.Annotate(users.Init, fx.As(new(users.Service))),
			fx.Annotate(oauthcodes.Init, fx.As(new(oauthcodes.Service))),
			fx.Annotate(oauthsessions.Init, fx.As(new(oauthsessions.Service))),
			fx.Annotate(oauthclients.Init, fx.As(new(oauthclients.Service))),
			fx.Annotate(oauthconsents.Init, fx.As(new(oauthconsents.Service))),
			fx.Annotate(websessions.Init, fx.As(new(websessions.Service))),
			fx.Annotate(oauth2.Init, fx.As(new(oauth2.Service))),
			fx.Annotate(inodes.Init, fx.As(new(inodes.Service))),
			fx.Annotate(files.Init, fx.As(new(files.Service))),
			fx.Annotate(davsessions.Init, fx.As(new(davsessions.Service))),
			fx.Annotate(folders.Init, fx.As(new(folders.Service))),

			// HTTP handlers
			AsRoute(dav.NewHTTPHandler),
			AsRoute(oauth2.NewHTTPHandler),
			AsRoute(assets.NewHTTPHandler),
			AsRoute(web.NewHTTPHandler),
			AsRoute(debug.NewHTTPHandler),

			// HTTP Router / HTTP Server
			router.InitMiddlewares,
			fx.Annotate(router.NewServer, fx.ParamTags(`group:"routes"`)),
		),

		// Start the command
		fx.Invoke(jobs.StartJobs),
		fx.Invoke(storage.RunMigrations),
		invoke,
	)

	return app
}
