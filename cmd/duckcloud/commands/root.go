package commands

import (
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/duckcloud/commands/bootstrap"
)

func NewServerCmd(binaryName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server <command>",
		Short: "Interact with your server",
	}

	cmd.AddCommand(NewRunCmd(binaryName))
	cmd.AddCommand(bootstrap.NewBootstrapCmd(binaryName))

	return cmd
}
