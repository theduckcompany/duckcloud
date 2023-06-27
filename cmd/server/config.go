package main

import (
	"github.com/Peltoche/neurone/src/tools/storage"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	Storage storage.Config
}
