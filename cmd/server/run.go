package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/config"
	"github.com/theduckcompany/duckcloud/src/server"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func NewRunCmd(binaryName string) *cobra.Command {
	var debug bool
	var dev bool

	fs := afero.NewOsFs()

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

			fmt.Printf("load database file from %q\n", cfg.Storage.Path)
			db, err := storage.Init(fs, &cfg.Storage, nil)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			server.Run(cfg, db, fs)
		},
	}

	cmd.Flags().BoolVar(&dev, "dev", false, "Run in dev mode and make json prettier")
	cmd.Flags().BoolVar(&debug, "debug", false, "Force the debug level")

	return &cmd
}
