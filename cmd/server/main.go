package main

import (
	"net/http"

	"github.com/Peltoche/neurone/src/service/dav"
	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"golang.org/x/exp/slog"
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

func main() {
	fx.New(
		fx.WithLogger(func(log *slog.Logger) fxevent.Logger { return logger.NewFxLogger(log) }),
		fx.Provide(
			NewDefaultConfig,

			storage.NewSQliteDBWithMigrate,
			logger.NewSLogger,
			fx.Annotate(tools.Init, fx.As(new(tools.Tools))),

			fx.Annotate(users.Init, fx.As(new(users.Service))),
			fx.Annotate(oauthclients.Init, fx.As(new(oauthclients.Service))),

			AsRoute(dav.NewHTTPHandler),
			AsRoute(users.NewHTTPHandler),

			fx.Annotate(
				router.NewChiRouter,
				fx.ParamTags(`group:"routes"`),
			),
			router.NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
