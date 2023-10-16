package bootstrap

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/config"
)

func setupAddr(cmd *cobra.Command, configSvc config.Service) {
	const (
		ServerWithProxy = "A server behind a local proxy (recommanded)"
		ServerExposed   = "A server directly exposed to internet"
		ServerDev       = "A local server for development (unsecure, hot-reload, pretty HTML, etc)"
	)

	const help = `This step is required to know how the server should expose it's data,
which security to enable, etc. You can find a detailed explanation of
each solutions below:

# Behind a local proxy (recommanded)

This solution is the most secure and the most flexible one for a production
server. The server will accept only the requests from "localhost" and will
exchange with the proxy in HTTP. In this scenario the proxy is responsible to
manage the HTTPS protocol between the client device and the server.

# Exposed directly on internet

This is the easiest way as it doesn't require a proxyand can be good solution for 
a quick test but it is not recommanded for a long living server as it would 
private all the other services from using the default HTTPS port.

This solution is only possibl with SSL enabled and there is two possibilites:
- Self-generate a SSL certificate: this is automatique but your website will 
  be marqued as unsafe unsafe by the browser because id doesn't know this 
	certificate.
- Bring your own signed certificated (a let's encrypt one for example).


# Localhost for dev (unsecure)

This solution is reserved for the local development. The server will accept the request
from everywhere in HTTP. It will also disable the HTML compression and will load the
HTTP templates for each calls instead of using the ones embedded in the binary. In this
situation the server will start only if it's run inside git repository's root directory.
`

	res, err := configSvc.GetAddrs(cmd.Context())
	switch {
	case err == nil:
		cmd.Printf("Addrs already setup: %q\n", res)
		return
	case errors.Is(err, config.ErrNotInitialized):
		// continue
	default:
		printErrAndExit(cmd, err)
	}

	prompt := &survey.Select{
		Message: `What kind of server do you want to install?`,
		Options: []string{ServerWithProxy, ServerExposed, ServerDev},
		Default: ServerWithProxy,
		Help:    help,
	}

	var template string
	var trustedHosts []string
	err = survey.AskOne(prompt, &template)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	var port int
	var hosts []string
	switch template {
	case ServerWithProxy:
		hosts = []string{"::1", "127.0.0.1"} // Accept the requests from localhost only
		port = askForPort(cmd, "7878")       // Some random port
		hostname := setupHostName(cmd, configSvc)
		disableSSL(cmd, configSvc)
		trustedHosts = []string{"::1", "127.0.0.1", "localhost", hostname}

	case ServerExposed:
		hosts = []string{"::"}        // Accept the requests from everywhere
		port = askForPort(cmd, "441") // HTTPS by default
		hostname := setupHostName(cmd, configSvc)
		trustedHosts = []string{"::1", "127.0.0.1", "localhost", hostname}
		enableSSL(cmd, configSvc)
	case ServerDev:
		hosts = []string{"::"}         // Accept the requests from everywhere
		port = askForPort(cmd, "8080") // Some classic dev port
		trustedHosts = []string{"::1", "127.0.0.1", "localhost"}

		err = configSvc.EnableDevMode(cmd.Context())
		if err != nil {
			printErrAndExit(cmd, fmt.Errorf("failed to enable dev_mode: %w", err))
		}

		err = configSvc.SetHostName(cmd.Context(), fmt.Sprintf("localhost:%d", port))
		if err != nil {
			printErrAndExit(cmd, fmt.Errorf("failed to set the hostname for dev: %w", err))
		}

		enableSSL(cmd, configSvc)
	default:
		printErrAndExit(cmd, errors.New("invalid selection"))
	}

	err = configSvc.SetAddrs(cmd.Context(), hosts, port)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	err = configSvc.SetTrustedHosts(cmd.Context(), trustedHosts)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to setup the trusted hosts: %w", err))
	}

	addrs, err := configSvc.GetAddrs(cmd.Context())
	if err != nil {
		printErrAndExit(cmd, err)
	}

	cmd.Printf("Addrs setup: %q\n", addrs)
	return
}

func askForPort(cmd *cobra.Command, defaultValue string) int {
	prompt := &survey.Input{
		Message: `What port do you want to use?`,
		Default: defaultValue,
	}

	var portStr string
	err := survey.AskOne(prompt, &portStr, survey.WithValidator(is.Port.Validate))
	if err != nil {
		printErrAndExit(cmd, err)
	}

	port, _ := strconv.Atoi(portStr) // Already validated above

	return port
}
