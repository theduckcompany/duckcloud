package server

import (
	"os"

	"github.com/Peltoche/neurone/src/server"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func NewStartCmd() *cobra.Command {
	var debug bool
	var dev bool

	cmd := cobra.Command{
		Short: "Start your server",
		Args:  cobra.NoArgs,
		Use:   "start",
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

			err := server.Start(cfg)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&dev, "dev", false, "Run in dev mode and make json prettier")
	cmd.Flags().BoolVar(&debug, "debug", false, "Force the debug level")

	return &cmd
}
