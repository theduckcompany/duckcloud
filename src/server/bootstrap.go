package server

import (
	"context"
	"fmt"

	"github.com/Peltoche/neurone/src/service/users"
	"go.uber.org/fx"
)

type BootstrapCmd struct {
	Username string
	Email    string
	Password string
}

func Bootstrap(ctx context.Context, cfg *Config, user users.CreateCmd) error {
	app := start(cfg, fx.Invoke(bootstrap(user)))

	err := app.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}

type bootstrapFunc = func(usersSvc users.Service) error

func bootstrap(cmd users.CreateCmd) bootstrapFunc {
	return func(usersSvc users.Service) error {
		ctx := context.Background()

		_, err := usersSvc.Create(ctx, &users.CreateCmd{
			Username: cmd.Username,
			Email:    cmd.Email,
			Password: cmd.Password,
		})
		if err != nil {
			return fmt.Errorf("failed to create the user: %w", err)
		}

		return nil
	}
}
