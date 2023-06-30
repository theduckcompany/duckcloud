package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Peltoche/neurone/src/service/assets"
	"github.com/Peltoche/neurone/src/service/dav"
	"github.com/Peltoche/neurone/src/service/oauth2"
	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
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

func Start(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	app := fx.New(
		fx.WithLogger(func(tools tools.Tools) fxevent.Logger { return logger.NewFxLogger(tools.Logger()) }),
		fx.Provide(
			func() Config { return *cfg },

			storage.NewSQliteDBWithMigrate,
			fx.Annotate(tools.NewToolbox, fx.As(new(tools.Tools))),

			fx.Annotate(users.Init, fx.As(new(users.Service))),
			fx.Annotate(oauthcodes.Init, fx.As(new(oauthcodes.Service))),
			fx.Annotate(oauthsessions.Init, fx.As(new(oauthsessions.Service))),
			fx.Annotate(oauthclients.Init, fx.As(new(oauthclients.Service))),

			AsRoute(dav.NewHTTPHandler),
			AsRoute(users.NewHTTPHandler),
			AsRoute(oauth2.NewHTTPHandler),
			AsRoute(assets.NewHTTPHandler),

			fx.Annotate(
				router.NewChiRouter,
				fx.ParamTags(`group:"routes"`),
			),
			router.NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	)

	if err := app.Err(); err != nil {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	app.Run()

	return nil
}
