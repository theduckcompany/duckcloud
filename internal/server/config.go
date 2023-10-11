package server

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/files"
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
	Listeners []router.Config `json:"listeners"`
	Assets    assets.Config   `json:"assets"`
	Storage   storage.Config  `json:"storage"`
	Files     files.Config    `json:"files"`
	Tools     tools.Config    `json:"tools"`
	Web       web.Config      `json:"web"`
}

func NewConfigFromDB(ctx context.Context, configSvc config.Service, folderPath string) (Config, error) {
	devModeEnabled, err := configSvc.IsDevModeEnabled(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check the dev mode: %w", err)
	}

	addrs, err := configSvc.Get(ctx, config.HTTPAddrs)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check the http addresses: %w", err)
	}

	isTLSEnabled, err := configSvc.IsTLSEnabled(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("failed to check if TLS is enabled: %w", err)
	}

	certifPath, err := configSvc.Get(ctx, config.SSLCertificatePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to retrieve the SSL certificate path: %w", err)
	}

	privateKeyPath, err := configSvc.Get(ctx, config.SSLPrivateKeyPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to retrieve the SSL private key path: %w", err)
	}

	// if debug {
	// 	cfg.Tools.Log.Level = slog.LevelDebug
	// }

	return Config{
		Listeners: []router.Config{
			{
				Addrs:    strings.Split(addrs, ","),
				TLS:      isTLSEnabled,
				CertFile: certifPath,
				KeyFile:  privateKeyPath,
				Services: []string{"dav", "auth", "assets", "web"},
			},
		},
		Assets: assets.Config{
			HotReload: devModeEnabled,
		},
		Files: files.Config{
			Path: path.Join(folderPath, "files"),
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
