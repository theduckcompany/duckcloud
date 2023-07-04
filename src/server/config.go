package server

import (
	"github.com/Peltoche/neurone/src/service/assets"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
	"golang.org/x/exp/slog"
)

type Config struct {
	fx.Out
	Listeners []router.Config `mapstructure:"listeners"`
	Assets    assets.Config   `mapstructure:"assets"`
	Storage   storage.Config  `mapstructure:"storage"`
	Tools     tools.Config    `mapstructure:"tools"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Listeners: []router.Config{
			{
				Port:          8080,
				TLS:           false,
				BindAddresses: []string{"::1", "127.0.0.1"},
				Services:      []string{"dav", "users", "auth", "assets"},
			},
		},
		Assets: assets.Config{
			HotReload: false,
		},
		Storage: storage.Config{
			Path: "./dev.db",
		},
		Tools: tools.Config{
			JWT: jwt.Config{
				Key: "A very bad key",
			},
			Response: response.Config{
				PrettyRender: false,
			},
			Log: logger.Config{
				Level: slog.LevelInfo,
			},
		},
	}
}
