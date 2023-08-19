package server

import (
	"log/slog"
	"path"

	"github.com/adrg/xdg"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"go.uber.org/fx"
)

const binaryName = "duckcloud"

type Config struct {
	fx.Out
	Listeners []router.Config `json:"listeners"`
	Assets    assets.Config   `json:"assets"`
	Storage   storage.Config  `json:"storage"`
	Files     files.Config    `json:"files"`
	Tools     tools.Config    `json:"tools"`
}

func NewDefaultConfig() *Config {
	dbPath, err := xdg.DataFile(path.Join(binaryName, "db.sqlite"))
	if err != nil {
		panic(err)
	}

	filesPath, err := xdg.DataFile(path.Join(binaryName, "files"))
	if err != nil {
		panic(err)
	}

	return &Config{
		Listeners: []router.Config{
			{
				Port:          8080,
				TLS:           false,
				BindAddresses: []string{"::1", "127.0.0.1"},
				Services:      []string{"dav", "users", "auth", "assets", "web"},
			},
		},
		Assets: assets.Config{
			HotReload: false,
		},
		Storage: storage.Config{
			Path:  dbPath,
			Debug: false,
		},
		Files: files.Config{
			Path: filesPath,
		},
		Tools: tools.Config{
			Response: response.Config{
				PrettyRender: false,
			},
			Log: logger.Config{
				Level: slog.LevelInfo,
			},
		},
	}
}
