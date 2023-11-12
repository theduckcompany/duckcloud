package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/cmd/duckcloud/commands"
	"github.com/theduckcompany/duckcloud/internal/tools/buildinfos"
)

const binaryName = "duckcloud"

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	cmd := &cobra.Command{
		Use:     binaryName,
		Short:   "Manage your duckcloud instance in your terminal.",
		Version: buildinfos.Version,
	}

	// Subcommands
	cmd.AddCommand(commands.NewRunCmd(binaryName))

	err := cmd.Execute()
	if err != nil {
		return exitError
	}

	return exitOK
}
