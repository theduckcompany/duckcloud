package server

import (
	"log/slog"
	"os"

	"github.com/myminicloud/myminicloud/cmd/config"
	"github.com/myminicloud/myminicloud/src/server"
	"github.com/spf13/cobra"
)

func NewRunCmd(binaryName string) *cobra.Command {
	var debug bool
	var dev bool

	cmd := cobra.Command{
		Short: "Run your server",
		Args:  cobra.NoArgs,
		Use:   "run",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg, err := config.GetOrCreateConfig(binaryName)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			if dev {
				cfg.Tools.Response.PrettyRender = true
				cfg.Tools.Response.HotReload = true
				cfg.Assets.HotReload = true
				cfg.Storage.Debug = true
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
