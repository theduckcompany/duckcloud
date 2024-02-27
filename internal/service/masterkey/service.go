package masterkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"golang.org/x/crypto/argon2"
)

var (
	ErrAlreadyExists        = errors.New("a master key already exists")
	ErrKeyAlreadyDeciphered = errors.New("the key have been already deciphered")
	ErrMasterKeyNotFound    = errors.New("master key not found")
)

const defaultPasswordKey = "p8ZY8JBIkK5qPvA4GLQkXVwY4fLDLPVkIvxOWy08DEs"

type PasswordSource string

type MasterKeyService struct {
	config  config.Service
	fs      afero.Fs
	enclave *memguard.Enclave
	cfg     Config

	passwordRequired bool
}

func NewService(config config.Service, fs afero.Fs, cfg Config) *MasterKeyService {
	return &MasterKeyService{
		config:  config,
		fs:      fs,
		enclave: nil,
		cfg:     cfg,

		passwordRequired: true,
	}
}

func (s *MasterKeyService) IsMasterKeyAvailable() bool {
	return s.enclave != nil
}

func (s *MasterKeyService) LoadMasterKeyFromPassword(ctx context.Context, password *secret.Text) error {
	if s.enclave != nil {
		return ErrKeyAlreadyDeciphered
	}

	masterKey, err := s.config.GetMasterKey(ctx)
	if errors.Is(err, errs.ErrNotFound) {
		return errs.NotFound(ErrMasterKeyNotFound)
	}

	if err != nil {
		return errs.Internal(fmt.Errorf("failed to get the master key: %w", err))
	}

	passKey, err := secret.KeyFromRaw(argon2.Key([]byte([]byte(password.Raw())), masterKey.Raw(), 3, 32*1024, 4, 32))
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to generate a passKey from the given password: %w", err))
	}

	rawMasterKey, err := masterKey.Open(passKey)
	if err != nil {
		return errs.BadRequest(fmt.Errorf("failed to decode: %w", err))
	}

	s.enclave = memguard.NewEnclave(rawMasterKey.Raw())

	return nil
}

func (s *MasterKeyService) loadMasterKeyFromSystemdCreds(ctx context.Context) error {
	password, err := s.loadPasswordFromSystemdCreds(s.fs)
	if err != nil {
		return fmt.Errorf("failed to load the systemd-creds password: %w", err)
	}

	return s.LoadMasterKeyFromPassword(ctx, password)
}

// func (s *MasterKeyService) generateMasterKey(ctx context.Context) error {
// 	existingMasterKey, err := s.config.GetMasterKey(ctx)
// 	if err != nil && !errors.Is(err, errs.ErrNotFound) {
// 		return fmt.Errorf("failed to get the master key: %w", err)
// 	}
//
// 	if existingMasterKey != nil {
// 		return nil
// 	}
//
// 	passwordKey, err := s.loadPassword(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to load the password: %w", err)
// 	}
//
// 	key, err := secret.NewKey()
// 	if err != nil {
// 		return fmt.Errorf("failed to generate a new key: %w", err)
// 	}
//
// 	sealedKey, err := secret.SealKey(passwordKey, key)
// 	if err != nil {
// 		return fmt.Errorf("failed to seal the key: %w", err)
// 	}
//
// 	err = s.config.SetMasterKey(ctx, sealedKey)
// 	if err != nil {
// 		return fmt.Errorf("failed to save into the storage: %w", err)
// 	}
//
// 	return nil
// }

func (s *MasterKeyService) loadPasswordFromSystemdCreds(fs afero.Fs) (*secret.Text, error) {
	dirPath := os.Getenv("CREDENTIALS_DIRECTORY")
	if dirPath == "" {
		return nil, errs.BadRequest(fmt.Errorf("CREDENTIALS_DIRECTORY not set"))
	}

	filePath := path.Join(dirPath, "password")

	file, err := fs.Open(filePath)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to open the credentials file specified by $CREDENTIALS_DIRECTORY: %w", err))
	}
	defer file.Close()

	password, err := io.ReadAll(file)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to read the password file: %w", err))
	}

	err = fs.Remove(filePath)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to remove the credentials file: %w", err))
	}

	passwordStr := secret.NewText(strings.TrimSpace(string(password)))

	return &passwordStr, nil
}

func (s *MasterKeyService) SealKey(key *secret.Key) (*secret.SealedKey, error) {
	sealedKey, err := secret.SealKeyWithEnclave(s.enclave, key)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to seal the key: %w", err))
	}

	return sealedKey, nil
}

func (s *MasterKeyService) Open(key *secret.SealedKey) (*secret.Key, error) {
	res, err := key.OpenWithEnclave(s.enclave)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}
