package main

import (
	"net/url"

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
	storageURL, _ := url.Parse("sqlite://./dev.db")
	return Config{
		Storage: storage.Config{
			URL: *storageURL,
		},
		JWT: jwt.Config{
			Key: "A very bad key",
		},
	}
}
