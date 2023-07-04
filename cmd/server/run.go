package server

import (
	"github.com/Peltoche/neurone/src/server"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func NewRunCmd() *cobra.Command {
	var debug bool
	var dev bool

	cmd := cobra.Command{
		Short: "Run your server",
		Args:  cobra.NoArgs,
		Use:   "run",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := server.NewDefaultConfig()

			if dev {
				cfg.Tools.Response.PrettyRender = true
				cfg.Tools.Response.HotReload = true
				cfg.Assets.HotReload = true
			}

			if debug {
				cfg.Tools.Log.Level = slog.LevelDebug
			}

			server.Run(cfg)
		},
	}

	cmd.Flags().BoolVar(&dev, "dev", false, "Run in dev mode and make json prettier")
	cmd.Flags().BoolVar(&debug, "debug", false, "Force the debug level")

	return &cmd
}
