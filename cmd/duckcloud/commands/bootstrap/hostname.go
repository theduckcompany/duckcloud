package bootstrap

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/config"
)

func setupHostName(cmd *cobra.Command, configSvc config.Service) string {
	const help = `The host name is the unique identifier that you will use to reach you server.

It can be either a domain name (like "cloud.mydomain.com") or an ip adresse (like "192.168.33").

This name will be used to generate you server url as follow: https://{{server-name}}/pictures`

	hostname, err := configSvc.Get(cmd.Context(), config.HostName)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	if hostname != "" {
		cmd.Printf("Hostname already setup: %q\n", hostname)
		return hostname
	}

	prompt := &survey.Input{
		Message: "What is your server host name?",
		Help:    help,
	}

	err = survey.AskOne(prompt, &hostname, survey.WithValidator(func(input interface{}) error {
		return configSvc.SetHostName(cmd.Context(), hostname)
	}))
	if err != nil {
		printErrAndExit(cmd, err)
	}

	cmd.Printf("Hostname setup: %q\n", hostname)

	return hostname
}
