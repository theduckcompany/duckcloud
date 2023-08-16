package server

import (
	"github.com/Peltoche/neurone/assets"
	"github.com/Peltoche/neurone/src/service/blocks"
	"github.com/Peltoche/neurone/src/service/dav"
	"github.com/Peltoche/neurone/src/service/fs"
	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/service/oauth2"
	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthconsents"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/web"
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
			fx.Annotate(blocks.Init, fx.As(new(blocks.Service))),

			// HTTP handlers
			AsRoute(dav.NewHTTPHandler),
			AsRoute(users.NewHTTPHandler),
			AsRoute(oauth2.NewHTTPHandler),
			AsRoute(assets.NewHTTPHandler),
			AsRoute(web.NewHTTPHandler),

			// HTTP Router / HTTP Server
			router.InitMiddlewares,
			fx.Annotate(router.NewServer, fx.ParamTags(`group:"routes"`)),
		),

		// Start the command
		fx.Invoke(fs.StartGC),
		fx.Invoke(storage.RunMigrations),
		invoke,
	)

	return app
}
