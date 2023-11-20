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
	GetKey(ctx context.Context, key ConfigKey) (*secret.SealedKey, error)
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
	key, err := s.storage.GetKey(ctx, masterKey)
	if errors.Is(err, errNotfound) {
		return nil, errs.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	return key, nil
}
