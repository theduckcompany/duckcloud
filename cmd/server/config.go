package main

import (
	"net/url"

	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	Storage storage.Config
}

func NewDefaultConfig() Config {
	storageURL, _ := url.Parse("sqlite://./dev.db")
	return Config{
		Storage: storage.Config{
			URL: *storageURL,
		},
	}
}
