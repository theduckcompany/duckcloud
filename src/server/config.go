package server

import (
	"log/slog"
	"path"

	"github.com/adrg/xdg"
	"github.com/myminicloud/myminicloud/assets"
	"github.com/myminicloud/myminicloud/src/service/blocks"
	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/logger"
	"github.com/myminicloud/myminicloud/src/tools/response"
	"github.com/myminicloud/myminicloud/src/tools/router"
	"github.com/myminicloud/myminicloud/src/tools/storage"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	Listeners []router.Config `json:"listeners"`
	Assets    assets.Config   `json:"assets"`
	Storage   storage.Config  `json:"storage"`
	Blocks    blocks.Config   `json:"blocks"`
	Tools     tools.Config    `json:"tools"`
}

func NewDefaultConfig() *Config {
	dbPath, err := xdg.DataFile(path.Join("myminicloud", "db.sqlite"))
	if err != nil {
		panic(err)
	}

	blocksPath, err := xdg.DataFile(path.Join("myminicloud", "blocks"))
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
		Blocks: blocks.Config{
			Path: blocksPath,
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
