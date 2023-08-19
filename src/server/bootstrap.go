package server

import (
	"context"
	"fmt"

	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"go.uber.org/fx"
)

type BootstrapCmd struct {
	Username string
	Email    string
	Password string
}

func Bootstrap(ctx context.Context, cfg *Config, user users.CreateCmd) error {
	//nolint:contextcheck // the bootstrap fonction must not use this context
	app := start(cfg, fx.Invoke(bootstrap(user)))

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

		user, err := usersSvc.Create(ctx, &users.CreateCmd{
			Username: cmd.Username,
			Email:    cmd.Email,
			Password: cmd.Password,
		})
		if err != nil {
			return fmt.Errorf("failed to create the user: %w", err)
		}

		_, err = clientSvc.Create(ctx, &oauthclients.CreateCmd{
			ID:             "web",
			Name:           "Web",
			RedirectURI:    "/settings",
			UserID:         string(user.ID()),
			Scopes:         []string{"*"},
			Public:         true,
			SkipValidation: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create the web client: %w", err)
		}

		return nil
	}
}
