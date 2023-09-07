package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/src/server"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

var qs = []*survey.Question{
	{
		Name: "username",
		Prompt: &survey.Input{
			Message: "What is your first user name?",
			Default: "admin",
		},
		Validate: survey.Required,
	},
	{
		Name:     "password",
		Prompt:   &survey.Password{Message: "Choose his password"},
		Validate: survey.Required,
	},
}

func NewBootstrapCmd(binaryName string) *cobra.Command {
	var debug bool

	fs := afero.NewOsFs()

	cmd := cobra.Command{
		Short: "Bootstrap your server",
		Args:  cobra.NoArgs,
		Use:   "bootstrap",
		Run: func(cmd *cobra.Command, _ []string) {
			answers := struct {
				Username string `survey:"username"`
				Password string `survey:"password"`
			}{}

			err := survey.Ask(qs, &answers)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			cfg := server.NewDefaultConfig()
			cfg.Tools.Log.Level = slog.LevelError

			if debug {
				cfg.Tools.Log.Level = slog.LevelDebug
			}

			bootCmd := users.CreateCmd{
				Username: answers.Username,
				Password: answers.Password,
			}

			fmt.Printf("load database file from %q\n", cfg.Storage.Path)
			db, err := storage.Init(fs, &cfg.Storage, nil)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			err = server.Bootstrap(cmd.Context(), db, fs, cfg, bootCmd)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			fmt.Println("Server successfully bootstraped!")
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "Force the debug level")

	return &cmd
}
