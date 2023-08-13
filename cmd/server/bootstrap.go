package server

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Peltoche/neurone/src/server"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
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
		Name:     "email",
		Prompt:   &survey.Input{Message: "Choose his email"},
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

	cmd := cobra.Command{
		Short: "Bootstrap your server",
		Args:  cobra.NoArgs,
		Use:   "bootstrap",
		Run: func(cmd *cobra.Command, _ []string) {
			answers := struct {
				Username string `survey:"username"`
				Email    string `survey:"email"`
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
				Email:    answers.Email,
				Password: answers.Password,
			}

			err = server.Bootstrap(cmd.Context(), cfg, bootCmd)
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
