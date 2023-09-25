package password

import (
	"context"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrMissmatchedPassword = bcrypt.ErrMismatchedHashAndPassword

type BcryptPassword struct{}

func NewBcryptPassword() *BcryptPassword {
	return &BcryptPassword{}
}

func (p *BcryptPassword) Encrypt(ctx context.Context, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hashedPassword), nil
}

func (p *BcryptPassword) Compare(ctx context.Context, hashStr, password string) error {
	rawPasswordHash, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return fmt.Errorf("failed to decode the password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(rawPasswordHash, []byte(password))

	return err
}
