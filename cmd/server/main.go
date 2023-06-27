package main

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/Peltoche/neurone/src/service/dav"
	"github.com/Peltoche/neurone/src/tools/httprouter"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
)

// AsMuxHandler annotates the given constructor to state that
// it provides a route to the "routes" group.
func AsMuxHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(httprouter.MuxHandler)),
		fx.ResultTags(`group:"routes"`),
	)
}

func main() {
	fx.New(
		// fx.WithLogger(func(log *logger.Logger) fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(
			func() Config {
				storageURL, _ := url.Parse("sqlite://./dev.db")
				return Config{
					Storage: storage.Config{
						URL: *storageURL,
					},
				}
			},
			logger.NewSLogger,
			storage.NewSQliteDBWithMigrate,

			AsMuxHandler(dav.NewHTTPHandler),

			fx.Annotate(
				httprouter.NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			httprouter.NewServer,
		),
		fx.Invoke(func(*http.Server, *sql.DB) {}),
	).Run()
}
