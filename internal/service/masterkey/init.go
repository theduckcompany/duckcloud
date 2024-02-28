package masterkey

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

var devMasterPassword = secret.NewText("duckpassword")

type Config struct {
	DevMode bool `mapstructure:"dev-mode"`
}

//go:generate mockery --name Service
type Service interface {
	LoadMasterKeyFromPassword(ctx context.Context, password *secret.Text) error
	IsMasterKeyLoaded() bool

	SealKey(key *secret.Key) (*secret.SealedKey, error)
	Open(key *secret.SealedKey) (*secret.Key, error)
}

func Init(ctx context.Context, config config.Service, fs afero.Fs, cfg Config, tools tools.Tools) (Service, error) {
	svc := NewService(config, fs, cfg)

	_, err := config.GetMasterKey(ctx)
	if errors.Is(err, errs.ErrNotFound) && cfg.DevMode {
		// Generate automatically the password with a default password at the moment.
		tools.Logger().Warn("No master key found and in dev mode: generate the default one")
		svc.generateMasterKey(ctx, &devMasterPassword)
	}

	err = svc.loadMasterKeyFromSystemdCreds(ctx)
	switch {
	case err == nil:
		return svc, nil
	case errors.Is(err, ErrCredsDirNotSet):
		tools.Logger().Warn("systemd-creds password not detected, needs to manualy set the password.")
		return svc, nil
	default:
		return nil, fmt.Errorf("master key error: %w", err)
	}
}
