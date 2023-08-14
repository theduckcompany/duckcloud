package server

import (
	"path"

	"github.com/Peltoche/neurone/assets"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/adrg/xdg"
	"go.uber.org/fx"
	"golang.org/x/exp/slog"
)

type Config struct {
	fx.Out
	Listeners []router.Config `json:"listeners"`
	Assets    assets.Config   `json:"assets"`
	Storage   storage.Config  `json:"storage"`
	Tools     tools.Config    `json:"tools"`
}

func NewDefaultConfig() *Config {
	dbPath, err := xdg.DataFile(path.Join("neurone", "db.sqlite"))
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
