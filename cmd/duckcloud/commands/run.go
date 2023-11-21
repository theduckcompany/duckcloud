package commands

import (
	"net"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/server"
	"github.com/theduckcompany/duckcloud/internal/tools/buildinfos"
)

var configDirs = append(xdg.DataDirs, xdg.DataHome)

func NewRunCmd(_ string) *cobra.Command {
	var defaultSpace string

	for _, dir := range configDirs {
		_, err := os.Stat(path.Join(dir, "duckcloud"))
		if err == nil {
			defaultSpace = path.Join(dir, "duckcloud")
			break
		}
	}

	if defaultSpace == "" {
		defaultSpace = path.Join(xdg.DataHome, "duckcloud")
	}

	cmd := cobra.Command{
		Short: "Run your server",
		Args:  cobra.NoArgs,
		Use:   "run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := NewConfigFromCmd(cmd)
			if err != nil {
				return err
			}

			server.Run(cmd.Context(), cfg)

			return nil
		},
	}

	flags := cmd.Flags()

	if !buildinfos.IsRelease() {
		// Those flags are  only available outside the releases for security reasons.
		flags.Bool("dev", false, "Run in dev mode and make json prettier")
		flags.Bool("hot-reload", false, "Enable the asset hot reload")
	}

	flags.Bool("debug", false, "Force the debug level")
	flags.String("log-level", "info", "Log message verbosity LEVEL (debug, info, warning, error)")

	flags.String("space", defaultSpace, "Specified you data directory location")
	flags.Bool("memory-fs", false, "Replace the OS filesystem by a in-memory stub. *Every data will disapear after each restart*.")

	flags.String("tls-cert", "", "Public HTTPS certificate FILE (.crt)")
	flags.String("tls-key", "", "Private HTTPS key FILE (.key)")
	flags.Bool("self-signed-cert", false, "Generate and use a self-signed HTTPS/TLS certificate ")

	flags.Int("http-port", 5764, "Web server port NUMBER, ignored for Unix domain sockets")
	flags.IP("http-host", net.IPv4(0, 0, 0, 0), "Web server IP address or Unix domain socket, e.g. unix:/var/run/photoprism.sock")
	flags.StringSlice("http-hostname", []string{}, "Serve requests for this HOSTNAME only plus")

	return &cmd
}
