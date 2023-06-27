package main

import (
	"net/http"

	"github.com/Peltoche/neurone/pkg/service/dav"
	"github.com/Peltoche/neurone/pkg/tools/httprouter"
	"github.com/Peltoche/neurone/pkg/tools/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
		fx.WithLogger(func(log *logger.Logger) fxevent.Logger { return fxevent.NopLogger }),
		fx.Provide(
			logger.NewSLogger,

			AsMuxHandler(dav.NewHTTPHandler),

			fx.Annotate(
				httprouter.NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			httprouter.NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
