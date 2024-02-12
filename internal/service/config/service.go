package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, key ConfigKey, value any) error
	Get(ctx context.Context, key ConfigKey, val any) error
}

type ConfigService struct {
	storage Storage
}

func NewService(storage Storage) *ConfigService {
	return &ConfigService{storage}
}

func (s *ConfigService) SetMasterKey(ctx context.Context, key *secret.SealedKey) error {
	err := s.storage.Save(ctx, masterKey, key)
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *ConfigService) GetMasterKey(ctx context.Context) (*secret.SealedKey, error) {
	var res secret.SealedKey

	err := s.storage.Get(ctx, masterKey, &res)
	if errors.Is(err, errNotfound) {
		return nil, errs.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return &res, nil
}
