package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/server"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

var configDirs = append([]string{xdg.DataHome}, xdg.DataDirs...)

func NewRunCmd(binaryName string) *cobra.Command {
	var debug bool
	var dev bool

	afs := afero.NewOsFs()

	cmd := cobra.Command{
		Short: "Run your server",
		Args:  cobra.NoArgs,
		Use:   "run",
		Run: func(cmd *cobra.Command, _ []string) {
			dataDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			if dataDir == "" {
				for _, dir := range configDirs {
					_, err := os.Stat(path.Join(dir, "duckcloud"))
					if err == nil {
						dataDir = path.Join(dir, "duckcloud")
						break
					}

					if !errors.Is(err, fs.ErrNotExist) {
						cmd.PrintErrln(err)
						os.Exit(1)
					}
				}

				if dataDir == "" {
					cmd.PrintErrln(fmt.Sprintf(`No data directory found, have you run "%s server bootstrap"?`, binaryName))
					os.Exit(1)
				}
			}

			dataDir, err = filepath.Abs(dataDir)
			if err != nil {
				cmd.PrintErrln(fmt.Sprintf(`invalid path %q: %s`, dataDir, err))
				os.Exit(1)
			}

			cmd.Printf("start server from: %s\n", dataDir)

			db, err := storage.NewSQliteClient(&storage.Config{
				Path:  path.Join(dataDir, "db.sqlite"),
				Debug: debug,
			}, logger.NewSLogger(logger.Config{Level: slog.LevelDebug}))
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			server.Run(cmd.Context(), db, afs)
		},
	}

	cmd.Flags().BoolVar(&dev, "dev", false, "Run in dev mode and make json prettier")
	cmd.Flags().BoolVar(&debug, "debug", false, "Force the debug level")

	return &cmd
}
