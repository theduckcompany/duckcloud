package bootstrap

import (
	"context"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/users"
)

type Config struct {
	Users   []users.CreateUserRequest `mapstructure:"users"`
	Clients []oauthclients.CreateCmd  `mapstructure:"clients"`
}

type Service interface {
	Bootstrap(ctx context.Context) error
}
