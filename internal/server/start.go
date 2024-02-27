package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/migrations"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/dav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/service/oauth2"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/stats"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/utilities"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tasks"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/cron"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/web"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/settings"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type Folder string

type Config struct {
	fx.Out
	Tools     tools.Config
	FS        afero.Fs
	Storage   storage.Config
	Folder    Folder
	Listener  router.Config
	HTML      html.Config
	Assets    assets.Config
	MasterKey masterkey.Config
}

// AsRoute annotates the given constructor to state that
// it provides a route to the "routes" group.
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(router.Registerer)),
		fx.ResultTags(`group:"routes"`),
	)
}

func start(ctx context.Context, cfg Config, invoke fx.Option) *fx.App {
	app := fx.New(
		fx.WithLogger(func(tools tools.Tools) fxevent.Logger { return logger.NewFxLogger(tools.Logger()) }),
		fx.Provide(
			func() context.Context { return ctx },
			func() Config { return cfg },

			func(folder Folder, fs afero.Fs, tools tools.Tools) (string, error) {
				folderPath, err := filepath.Abs(string(folder))
				if err != nil {
					return "", fmt.Errorf("invalid path: %q: %w", folderPath, err)
				}

				err = fs.MkdirAll(string(folder), 0o755)
				if err != nil && !errors.Is(err, os.ErrExist) {
					return "", fmt.Errorf("failed to create the %s: %w", folderPath, err)
				}

				if fs.Name() == afero.NewMemMapFs().Name() {
					tools.Logger().Info("Load data from memory")
				} else {
					tools.Logger().Info(fmt.Sprintf("Load data from %s", folder))
				}

				return folderPath, nil
			},

			// Tools
			fx.Annotate(tools.NewToolbox, fx.As(new(tools.Tools))),
			fx.Annotate(html.NewRenderer, fx.As(new(html.Writer))),
			storage.Init,
			auth.NewAuthenticator,

			// Services
			users.Init,
			dfs.Init,
			files.Init,
			fx.Annotate(config.Init, fx.As(new(config.Service))),
			fx.Annotate(oauthcodes.Init, fx.As(new(oauthcodes.Service))),
			fx.Annotate(oauthsessions.Init, fx.As(new(oauthsessions.Service))),
			fx.Annotate(oauthclients.Init, fx.As(new(oauthclients.Service))),
			fx.Annotate(oauthconsents.Init, fx.As(new(oauthconsents.Service))),
			fx.Annotate(websessions.Init, fx.As(new(websessions.Service))),
			fx.Annotate(oauth2.Init, fx.As(new(oauth2.Service))),
			fx.Annotate(davsessions.Init, fx.As(new(davsessions.Service))),
			fx.Annotate(spaces.Init, fx.As(new(spaces.Service))),
			fx.Annotate(scheduler.Init, fx.As(new(scheduler.Service))),
			fx.Annotate(stats.Init, fx.As(new(stats.Service))),
			fx.Annotate(masterkey.Init, fx.As(new(masterkey.Service))),

			// Tasks
			tasks.Init,

			// HTTP handlers
			AsRoute(dav.NewHTTPHandler),
			AsRoute(oauth2.NewHTTPHandler),
			AsRoute(assets.NewHTTPHandler),
			AsRoute(utilities.NewHTTPHandler),

			// Web Pages
			// AsRoute(web.NewHTTPHandler),
			AsRoute(web.NewHomePage),
			AsRoute(auth.NewLoginPage),
			AsRoute(auth.NewConsentPage),
			AsRoute(settings.NewHandler),
			AsRoute(settings.NewSecurityPage),
			AsRoute(settings.NewSpacesPage),
			AsRoute(settings.NewUsersPage),

			// HTTP Router / HTTP Server
			router.InitMiddlewares,
			fx.Annotate(router.NewServer, fx.ParamTags(`group:"routes"`)),

			// Task Runner
			fx.Annotate(runner.Init, fx.ParamTags(`group:"tasks"`), fx.As(new(runner.Service))),
		),

		fx.Invoke(migrations.Run),

		fx.Invoke(bootstrap),

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

		fx.Invoke(func(ctx context.Context, runner runner.Service) error {
			return runner.Run(ctx)
		}),

		invoke,
	)

	return app
}
