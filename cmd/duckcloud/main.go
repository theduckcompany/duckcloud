package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/duckcloud/commands"
	"github.com/theduckcompany/duckcloud/cmd/duckcloud/commands/bootstrap"
	"github.com/theduckcompany/duckcloud/internal/tools/buildinfos"
)

const binaryName = "duckcloud"

func main() {
	cmd := &cobra.Command{
		Use:     binaryName,
		Short:   "Manage your duckcloud instance in your terminal.",
		Version: buildinfos.Version,
	}

	// Generic flags
	cmd.PersistentFlags().StringP("dir", "d", "", "Specified you data directory location")

	// Subcommands
	cmd.AddCommand(commands.NewRunCmd(binaryName))
	cmd.AddCommand(bootstrap.NewBootstrapCmd(binaryName))

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
