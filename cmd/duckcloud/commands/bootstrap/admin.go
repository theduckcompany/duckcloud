package bootstrap

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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

func setupAdmin(cmd *cobra.Command, userSvc users.Service) {
	answers := struct {
		Username string `survey:"username"`
		Password string `survey:"password"`
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	user, err := userSvc.Create(cmd.Context(), &users.CreateCmd{
		Username: answers.Username,
		Password: answers.Password,
	})
	if err != nil {
		printErrAndExit(cmd, err)
	}

	fmt.Printf("Server %q successfully bootstraped!\n", user.Username())
}
