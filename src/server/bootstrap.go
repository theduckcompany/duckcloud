package server

import (
	"context"
	"fmt"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/users"
	"go.uber.org/fx"
)

type BootstrapCmd struct {
	Username string
	Email    string
	Password string
}

func Bootstrap(ctx context.Context, cfg *Config, user users.CreateUserRequest) error {
	app := start(cfg, fx.Invoke(bootstrap(user)))

	err := app.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}

type bootstrapFunc = func(usersSvc users.Service, oauthClients oauthclients.Service) error

func bootstrap(cmd users.CreateUserRequest) bootstrapFunc {
	return func(usersSvc users.Service, oauthClients oauthclients.Service) error {
		ctx := context.Background()

		user, err := usersSvc.Create(ctx, &users.CreateUserRequest{
			Username: cmd.Username,
			Email:    cmd.Email,
			Password: cmd.Password,
		})
		if err != nil {
			return fmt.Errorf("failed to create the user: %w", err)
		}

		err = oauthClients.Create(ctx, &oauthclients.CreateCmd{
			ID:             oauthclients.WebAppClientID,
			Name:           "Neurone Web App",
			RedirectURI:    "http://localhost:8080",
			UserID:         string(user.ID),
			Scopes:         oauthclients.Scopes{"users"},
			Public:         true,
			SkipValidation: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create the web app client: %w", err)
		}

		return nil
	}
}
