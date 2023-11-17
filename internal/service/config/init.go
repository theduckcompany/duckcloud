package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

//go:generate mockery --name Service
type Service interface {
	SetMasterKey(ctx context.Context, key *secret.Key) error
	GetMasterKey(ctx context.Context) (*secret.Key, error)
}

func Init(db *sql.DB) (Service, error) {
	storage := newSqlStorage(db)

	svc := NewService(storage)

	masterKey, err := svc.GetMasterKey(context.Background())
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("failed to get the master key: %w", err)
	}

	if masterKey == nil {
		err = svc.generateMasterKey(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to generate a new master key: %w", err)
		}
	}

	return svc, nil
}
