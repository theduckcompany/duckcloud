package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/response"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/web"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	Listeners []router.Config
	Assets    assets.Config
	Storage   storage.Config
	Tools     tools.Config
	Web       web.Config
}

func NewConfigFromDB(ctx context.Context, configSvc config.Service, folderPath string) (Config, error) {
	devModeEnabled, err := configSvc.IsDevModeEnabled(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check the dev mode: %w", err)
	}

	addrs, err := configSvc.GetAddrs(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check the http addresses: %w", err)
	}

	isTLSEnabled, err := configSvc.IsTLSEnabled(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check if TLS is enabled: %w", err)
	}

	certifPath, privateKeyPath, err := configSvc.GetSSLPaths(ctx)
	if err != nil && !errors.Is(err, config.ErrNotInitialized) {
		return Config{}, fmt.Errorf("failed to retrieve the SSL paths: %w", err)
	}

	return Config{
		Listeners: []router.Config{
			{
				Addrs:    addrs,
				TLS:      isTLSEnabled,
				CertFile: certifPath,
				KeyFile:  privateKeyPath,
				Services: []string{"dav", "auth", "assets", "web"},
			},
		},
		Assets: assets.Config{
			HotReload: devModeEnabled,
		},
		Tools: tools.Config{
			Response: response.Config{
				PrettyRender: devModeEnabled,
			},
			Log: logger.Config{
				Level: slog.LevelInfo,
			},
		},
		Web: web.Config{
			HTML: html.Config{
				PrettyRender: devModeEnabled,
				HotReload:    devModeEnabled,
			},
		},
	}, nil
}
