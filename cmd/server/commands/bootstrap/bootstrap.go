package bootstrap

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/response"
)

func printErrAndExit(cmd *cobra.Command, err error) {
	cmd.PrintErrln(err)
	os.Exit(1)
}

func NewBootstrapCmd(_ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap your server",
		Run: func(cmd *cobra.Command, _ []string) {
			dataDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			folderPath := bootstrapFolder(cmd, dataDir)
			db := bootstrapDB(cmd, folderPath)

			configSvc := config.Init(db)

			setupFolderPath(cmd, configSvc, folderPath)
			setupAddr(cmd, configSvc)
			setupSSLCertificate(cmd, configSvc)

			tools := tools.NewToolbox(tools.Config{
				Response: response.Config{},
				Log:      logger.Config{Level: slog.LevelInfo},
			})
			inodeSvc := inodes.Init(tools, db)
			folderSvc := folders.Init(tools, db, inodeSvc)
			userSvc := users.Init(tools, db, folderSvc)

			setupAdmin(cmd, userSvc)
		},
	}

	return cmd
}
