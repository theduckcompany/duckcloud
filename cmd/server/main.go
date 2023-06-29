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
	var dev bool

	flag.BoolVar(&dev, "dev", false, "Run in dev mode and make json prettier")
	flag.BoolVar(&debug, "debug", false, "Force the debug level")

	flag.Parse()

	if dev {
		cfg.Tools.Response.PrettyRender = true
		cfg.Tools.Response.HotReload = true
	}

	if debug {
		cfg.Tools.Log.Level = slog.LevelDebug
	}

	err := server.Start(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
