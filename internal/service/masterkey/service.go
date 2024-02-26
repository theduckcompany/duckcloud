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

	masterKeyAvailable bool
	passwordRequired   bool
}

func NewService(config config.Service, fs afero.Fs, cfg Config) *MasterKeyService {
	return &MasterKeyService{
		config:  config,
		fs:      fs,
		enclave: nil,
		cfg:     cfg,

		masterKeyAvailable: false,
		passwordRequired:   true,
	}
}

func (s *MasterKeyService) IsMasterKeyAvailable() bool {
	return s.masterKeyAvailable
}

func (s *MasterKeyService) loadMasterKey(ctx context.Context) error {
	masterKey, err := s.config.GetMasterKey(ctx)
	if errors.Is(err, errs.ErrNotFound) {
		s.masterKeyAvailable = false
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to get the master key: %w", err)
	}

	passwordKey, err := s.loadPassword(ctx)
	if err != nil {
		return fmt.Errorf("failed to load password: %w", err)
	}

	if passwordKey == nil {
		s.passwordRequired = true
		return nil
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
	switch {
	case s.cfg.DevMode:
		return secret.KeyFromBase64(defaultPasswordKey)
	case os.Getenv("CREDENTIALS_DIRECTORY") != "":
		return s.loadPasswordFromSystemdCreds(ctx, s.fs)
	default:
		return nil, nil
	}
}

func (s *MasterKeyService) loadPasswordFromSystemdCreds(ctx context.Context, fs afero.Fs) (*secret.Key, error) {
	dirPath := os.Getenv("CREDENTIALS_DIRECTORY")
	if dirPath == "" {
		return nil, fmt.Errorf("systemd credentials: %w", errs.ErrNotFound)
	}

	filePath := path.Join(dirPath, "password")

	file, err := fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the credentials file specified by $CREDENTIALS_DIRECTORY: %w", err)
	}
	defer file.Close()

	password, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the password file: %w", err)
	}

	err = fs.Remove(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to remove the credentials file: %w", err)
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
