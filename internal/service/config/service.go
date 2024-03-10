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
	Save(ctx context.Context, key ConfigKey, value string) error
	Get(ctx context.Context, key ConfigKey) (string, error)
}

type service struct {
	storage Storage
}

func newService(storage Storage) *service {
	return &service{storage}
}

func (s *service) SetMasterKey(ctx context.Context, key *secret.SealedKey) error {
	err := s.storage.Save(ctx, masterKey, key.Base64())
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *service) GetMasterKey(ctx context.Context) (*secret.SealedKey, error) {
	keyStr, err := s.storage.Get(ctx, masterKey)
	if errors.Is(err, errNotfound) {
		return nil, errs.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	res, err := secret.SealedKeyFromBase64(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the key: %w", err)
	}

	return res, nil
}
