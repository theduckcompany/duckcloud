package masterkey

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

type Config struct {
	DevMode bool `mapstructure:"dev-mode"`
}

//go:generate mockery --name Service
type Service interface {
	SealKey(key *secret.Key) (*secret.SealedKey, error)
	Open(key *secret.SealedKey) (*secret.Key, error)
}

func Init(ctx context.Context, config config.Service, fs afero.Fs, cfg Config) (Service, error) {
	svc := NewService(config, fs, cfg)

	err := svc.generateMasterKey(ctx)
	if err != nil {
		return nil, err
	}

	err = svc.loadMasterKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("master key error: %w", err)
	}

	return svc, nil
}
