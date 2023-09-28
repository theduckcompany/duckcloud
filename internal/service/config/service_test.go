package config

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("Bootstrap", func(t *testing.T) {
		tests := []struct {
			Name     string
			Cmd      BootstrapCmd
			Expected error
		}{
			{
				Name:     "Success an IP as host",
				Cmd:      BootstrapCmd{"127.0.0.1", 80},
				Expected: nil,
			},
			{
				Name:     "Success with a IP:port as host",
				Cmd:      BootstrapCmd{"127.0.0.1:441", 80},
				Expected: nil,
			},
			{
				Name:     "Success with a domain as host",
				Cmd:      BootstrapCmd{"example.com", 80},
				Expected: nil,
			},
			{
				Name:     "Success with a domain:port as host",
				Cmd:      BootstrapCmd{"example.com:80", 80},
				Expected: nil,
			},
			{
				Name:     "Success with a sub.domain:port as host",
				Cmd:      BootstrapCmd{"foo.bar.tutu.example.com:9090", 1234},
				Expected: nil,
			},
			{
				Name:     "With a port too big",
				Cmd:      BootstrapCmd{"foo.bar.tutu.example.com:98090", 1234},
				Expected: errors.New("invalid hostname: must be a valid port number"),
			},
			{
				Name:     "With an invalid port for the host",
				Cmd:      BootstrapCmd{"example.com:fr32", 1234},
				Expected: errors.New("invalid hostname: must be a valid port number"),
			},
			{
				Name:     "With to many separator",
				Cmd:      BootstrapCmd{"example.com:32:32", 1234},
				Expected: errors.New("invalid hostname: address example.com:32:32: too many colons in address"),
			},
			{
				Name:     "With an invalid host",
				Cmd:      BootstrapCmd{"e&x#a@mple.com:32", 1234},
				Expected: errors.New("invalid hostname: must be a valid IP address or DNS name"),
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				// TODO: use a noop driver. At the moment we need to create a new
				// db and run the complete migration for each test even if we never
				// check the results.
				db := storage.NewTestStorage(t)
				storage := newSqlStorage(db)
				svc := NewService(storage)

				err := svc.Bootstrap(ctx, test.Cmd)
				if test.Expected == nil {
					assert.NoError(t, err, test.Cmd)
				} else {
					require.Error(t, err, "expected error %q for input %#v", test.Expected, test.Cmd)
					assert.EqualError(t, test.Expected, err.Error(), test.Cmd)
				}
			})
		}
	})
}
