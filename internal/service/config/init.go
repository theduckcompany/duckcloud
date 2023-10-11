package config

import (
	"context"
	"database/sql"
)

//go:generate mockery --name Service
type Service interface {
	SetHostName(ctx context.Context, hostName string) error
	SetAddrs(ctx context.Context, hosts []string, port int) error
	SetTrustedHosts(ctx context.Context, hosts []string) error
	EnableTLS(ctx context.Context) error
	EnableDevMode(ctx context.Context) error
	SetSSLPaths(ctx context.Context, certifPath, privateKeyPath string) error
	DisableTLS(ctx context.Context) error
	IsTLSEnabled(ctx context.Context) (bool, error)
	IsDevModeEnabled(ctx context.Context) (bool, error)
	GetTrustedHosts(ctx context.Context) ([]string, error)
	GetSSLPaths(ctx context.Context) (string, string, error)
	GetAddrs(ctx context.Context) ([]string, error)
	GetHostName(ctx context.Context) (string, error)
}

func Init(db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(storage)
}
