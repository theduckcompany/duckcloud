package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/server/commands"
)

const binaryName = "duckcloud"

func main() {
	cmd := &cobra.Command{
		Use:   binaryName,
		Short: "Manage your duckcloud instance in your terminal.",
	}

	// Generic flags

	// tb := toolbox.NewProd()

	// Subcommands
	cmd.AddCommand(commands.NewServerCmd(binaryName))

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}