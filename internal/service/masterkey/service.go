package masterkey

import (
	"context"
	"encoding/hex"
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
)

var ErrAlreadyExists = errors.New("a master key already exists")

const defaultPasswordKey = "p8ZY8JBIkK5qPvA4GLQkXVwY4fLDLPVkIvxOWy08DEs"

type PasswordSource string

type MasterKeyService struct {
	config  config.Service
	fs      afero.Fs
	enclave *memguard.Enclave
	cfg     Config
}

func NewService(config config.Service, fs afero.Fs, cfg Config) *MasterKeyService {
	return &MasterKeyService{config, fs, nil, cfg}
}

func (s *MasterKeyService) loadMasterKey(ctx context.Context) error {
	masterKey, err := s.config.GetMasterKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get the master key: %w", err)
	}

	passwordKey, err := s.loadPassword(ctx)
	if err != nil {
		return fmt.Errorf("failed to load password: %w", err)
	}

	rawMasterKey, err := masterKey.Open(passwordKey)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	s.enclave = memguard.NewEnclave(rawMasterKey.Raw())

	return nil
}

func (s *MasterKeyService) generateMasterKey(ctx context.Context) error {
	existingMasterKey, err := s.config.GetMasterKey(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("failed to get the master key: %w", err)
	}

	if existingMasterKey != nil {
		return nil
	}

	passwordKey, err := s.loadPassword(ctx)
	if err != nil {
		return fmt.Errorf("failed to load the password: %w", err)
	}

	key, err := secret.NewKey()
	if err != nil {
		return fmt.Errorf("failed to generate a new key: %w", err)
	}

	sealedKey, err := secret.SealKey(passwordKey, key)
	if err != nil {
		return fmt.Errorf("failed to seal the key: %w", err)
	}

	err = s.config.SetMasterKey(ctx, sealedKey)
	if err != nil {
		return fmt.Errorf("failed to save into the storage: %w", err)
	}

	return nil
}

func (s *MasterKeyService) loadPassword(ctx context.Context) (*secret.Key, error) {
	if s.cfg.DevMode {
		passwordKey, err := secret.KeyFromBase64(defaultPasswordKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load the default password key: %w", err)
		}

		return passwordKey, nil
	}

	passwordKey, err := s.loadPasswordFromSystemdCreds(ctx, s.fs)
	if err != nil {
		return nil, fmt.Errorf("failed to load password: %w", err)
	}

	return passwordKey, nil
}

func (s *MasterKeyService) loadPasswordFromSystemdCreds(ctx context.Context, fs afero.Fs) (*secret.Key, error) {
	dirPath := os.Getenv("CREDENTIALS_DIRECTORY")
	if dirPath == "" {
		return nil, fmt.Errorf("systemd credentials: %w", errs.ErrNotFound)
	}

	file, err := fs.Open(path.Join(dirPath, "password"))
	if err != nil {
		return nil, fmt.Errorf("failed to open the credentials file specified by $CREDENTIALS_DIRECTORY: %w", err)
	}
	defer file.Close()

	password, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the password file: %w", err)
	}

	passwordStr := strings.TrimSuffix(string(password), "\n")

	pass, err := hex.DecodeString(passwordStr)
	if err != nil {
		return nil, fmt.Errorf("decode 3 error: %w", err)
	}

	passwordKey, err := secret.KeyFromRaw(pass)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the user password: %w", err)
	}

	return passwordKey, nil
}

func (s *MasterKeyService) SealKey(key *secret.Key) (*secret.SealedKey, error) {
	sealedKey, err := secret.SealKeyWithEnclave(s.enclave, key)
	if err != nil {
		return nil, fmt.Errorf("failed to seal the key: %w", err)
	}

	return sealedKey, nil
}

func (s *MasterKeyService) Open(key *secret.SealedKey) (*secret.Key, error) {
	return key.OpenWithEnclave(s.enclave)
}
