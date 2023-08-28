package server

import (
	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/src/jobs"
	"github.com/theduckcompany/duckcloud/src/service/dav"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/debug"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/oauth2"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/web"
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

func start(cfg *Config, invoke fx.Option) *fx.App {
	app := fx.New(
		fx.WithLogger(func(tools tools.Tools) fxevent.Logger { return logger.NewFxLogger(tools.Logger()) }),
		fx.Provide(
			func() Config { return *cfg },
			afero.NewOsFs,

			// Tools
			storage.Init,
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

			// HTTP handlers
			AsRoute(dav.NewHTTPHandler),
			AsRoute(users.NewHTTPHandler),
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
