package server

import (
	"context"
	"database/sql"
	"time"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/dav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/debug"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauth2"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/cron"
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

// AsTask annotates the given constructor to state that
// it provides a task to the "tasks" group.
func AsTask(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(runner.TaskRunner)),
		fx.ResultTags(`group:"tasks"`),
	)
}

func start(ctx context.Context, db *sql.DB, fs afero.Fs, folderPath string, invoke fx.Option) *fx.App {
	app := fx.New(
		fx.WithLogger(func(tools tools.Tools) fxevent.Logger { return logger.NewFxLogger(tools.Logger()) }),
		fx.Provide(
			func() string { return folderPath },
			func() context.Context { return ctx },
			func() afero.Fs { return fs },
			func() *sql.DB { return db },
			NewConfigFromDB,

			// Tools
			fx.Annotate(tools.NewToolbox, fx.As(new(tools.Tools))),

			// Services
			users.Init,
			dfs.Init,
			fx.Annotate(config.Init, fx.As(new(config.Service))),
			fx.Annotate(oauthcodes.Init, fx.As(new(oauthcodes.Service))),
			fx.Annotate(oauthsessions.Init, fx.As(new(oauthsessions.Service))),
			fx.Annotate(oauthclients.Init, fx.As(new(oauthclients.Service))),
			fx.Annotate(oauthconsents.Init, fx.As(new(oauthconsents.Service))),
			fx.Annotate(websessions.Init, fx.As(new(websessions.Service))),
			fx.Annotate(oauth2.Init, fx.As(new(oauth2.Service))),
			fx.Annotate(davsessions.Init, fx.As(new(davsessions.Service))),
			fx.Annotate(folders.Init, fx.As(new(folders.Service))),
			fx.Annotate(scheduler.Init, fx.As(new(scheduler.Service))),

			// HTTP handlers
			AsRoute(dav.NewHTTPHandler),
			AsRoute(oauth2.NewHTTPHandler),
			AsRoute(assets.NewHTTPHandler),
			AsRoute(web.NewHTTPHandler),
			AsRoute(debug.NewHTTPHandler),

			// HTTP Router / HTTP Server
			router.InitMiddlewares,
			fx.Annotate(router.NewServer, fx.ParamTags(`group:"routes"`)),

			// Task Runner
			fx.Annotate(runner.Init, fx.ParamTags(`group:"tasks"`), fx.As(new(runner.Service))),
		),

		// Run the migration
		fx.Invoke(storage.RunMigrations),

		// Start the tasks-runner
		fx.Invoke(func(svc runner.Service, lc fx.Lifecycle, tools tools.Tools) {
			cronSvc := cron.New("tasks-runner", 500*time.Millisecond, tools, svc)
			cronSvc.FXRegister(lc)
		}),

		// Start the scheduler
		fx.Invoke(func(svc scheduler.Service, lc fx.Lifecycle, tools tools.Tools) {
			cronSvc := cron.New("tasks-scheduler", 10*time.Second, tools, svc)
			cronSvc.FXRegister(lc)
		}),
		invoke,
	)

	return app
}
