package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"go.uber.org/fx"
)

type BootstrapCmd struct {
	Username string
	Password string
}

func Bootstrap(ctx context.Context, db *sql.DB, fs afero.Fs, cfg *Config, user users.CreateCmd) error {
	//nolint:contextcheck // the bootstrap fonction must not use this context
	app := start(cfg, db, fs, fx.Invoke(bootstrap(user)))

	err := app.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	return nil
}

type bootstrapFunc = func(usersSvc users.Service, clientSvc oauthclients.Service) error

func bootstrap(cmd users.CreateCmd) bootstrapFunc {
	return func(usersSvc users.Service, clientSvc oauthclients.Service) error {
		ctx := context.Background()

		_, err := usersSvc.Create(ctx, &users.CreateCmd{
			Username: cmd.Username,
			Password: cmd.Password,
			IsAdmin:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to create the user: %w", err)
		}

		return nil
	}
}
