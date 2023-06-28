package main

import (
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	Storage storage.Config `mapstructure:"storage"`
	JWT     jwt.Config     `mapstructure:"jwt"`
}

func NewDefaultConfig() Config {
	return Config{
		Storage: storage.Config{
			Path: "./dev.db",
		},
		JWT: jwt.Config{
			Key: "A very bad key",
		},
	}
}
