package server

import "github.com/spf13/cobra"

func NewServerCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "server <command>",
		Short: "Interact with your server",
	}

	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewBootstrapCmd())

	return cmd
}
