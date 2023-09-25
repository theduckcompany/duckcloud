package commands

import "github.com/spf13/cobra"

func NewServerCmd(binaryName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server <command>",
		Short: "Interact with your server",
	}

	cmd.AddCommand(NewRunCmd(binaryName))
	cmd.AddCommand(NewBootstrapCmd(binaryName))

	return cmd
}
