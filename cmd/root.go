package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/server"
)

const binaryName = "duckcloud"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := &cobra.Command{
		Use:   binaryName,
		Short: "Manage your duckcloud instance in your terminal.",
	}

	// Generic flags

	// tb := toolbox.NewProd()

	// Subcommands
	cmd.AddCommand(server.NewServerCmd(binaryName))

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
