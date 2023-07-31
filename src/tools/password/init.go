package password

import "context"

//go:generate mockery --name Password
type Password interface {
	Encrypt(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, hash, password string) error
}
