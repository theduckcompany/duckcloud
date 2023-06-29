package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Peltoche/neurone/src/server"
	"golang.org/x/exp/slog"
)

func main() {
	cfg := server.NewDefaultConfig()

	var debug bool

	flag.BoolVar(&cfg.Dev, "dev", false, "Run in dev mode and make json prettier")
	flag.BoolVar(&debug, "debug", false, "Force the debug level")

	flag.Parse()

	if debug {
		cfg.Log.Level = slog.LevelDebug
	}

	err := server.Start(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
