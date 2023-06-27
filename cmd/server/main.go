package main

import (
	"database/sql"
	"net/http"

	"github.com/Peltoche/neurone/src/service/dav"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
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
		// fx.WithLogger(func(log *logger.Logger) fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(
			NewDefaultConfig,

			logger.NewSLogger,
			storage.NewSQliteDBWithMigrate,

			AsRoute(dav.NewHTTPHandler),

			fx.Annotate(
				router.NewChiRouter,
				fx.ParamTags(`group:"routes"`),
			),
			router.NewServer,
		),
		fx.Invoke(func(*http.Server, *sql.DB) {}),
	).Run()
}
