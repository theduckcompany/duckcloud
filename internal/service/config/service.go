package config

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var (
	ErrAlreadyBootstraped = errors.New("server already bootstraped")
	ErrInvalidHostname    = errors.New("invalid hostname")
	ErrInvalidPort        = errors.New("invalid port")
	ErrSSLMustBeEnabled   = errors.New("TLS must be enabled")
	ErrInvalidPEMFormat   = errors.New("invalid PEM format")
	ErrMustNotHavePort    = errors.New("must not have port")
	ErrNotInitialized     = errors.New("not initialized")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, key ConfigKey, value string) error
	Get(ctx context.Context, key ConfigKey) (string, error)
}

type ConfigService struct {
	storage Storage
}

func NewService(storage Storage) *ConfigService {
	return &ConfigService{storage}
}

func (s *ConfigService) EnableTLS(ctx context.Context) error {
	return s.storage.Save(ctx, tlsEnabled, "true")
}

func (s *ConfigService) EnableDevMode(ctx context.Context) error {
	return s.storage.Save(ctx, devModeEnabled, "true")
}

func (s *ConfigService) IsDevModeEnabled(ctx context.Context) (bool, error) {
	return s.checkBool(ctx, devModeEnabled)
}

func (s *ConfigService) DisableTLS(ctx context.Context) error {
	return s.storage.Save(ctx, tlsEnabled, "false")
}

func (s *ConfigService) IsTLSEnabled(ctx context.Context) (bool, error) {
	return s.checkBool(ctx, tlsEnabled)
}

func (s *ConfigService) checkBool(ctx context.Context, key ConfigKey) (bool, error) {
	enabled, err := s.storage.Get(ctx, key)
	if errors.Is(err, errNotfound) {
		return false, ErrNotInitialized
	}

	if err != nil {
		return false, fmt.Errorf("storage error: %w", err)
	}

	return enabled == "true", nil
}

func (s *ConfigService) SetSSLPaths(ctx context.Context, certifPath, privateKeyPath string) error {
	enabled, err := s.IsTLSEnabled(ctx)
	if err != nil {
		return err
	}

	if !enabled {
		return ErrSSLMustBeEnabled
	}

	err = s.setPEMPath(ctx, sslCertificatePath, certifPath)
	if err != nil {
		return fmt.Errorf("failed to save the certificate: %w", err)
	}

	err = s.setPEMPath(ctx, sslPrivateKeyPath, privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to save the private key: %w", err)
	}

	return nil
}

func (s *ConfigService) GetSSLPaths(ctx context.Context) (string, string, error) {
	certif, err := s.storage.Get(ctx, sslCertificatePath)
	if errors.Is(err, errNotfound) {
		return "", "", ErrNotInitialized
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to fetch the SSL certificate: %w", err)
	}

	key, err := s.storage.Get(ctx, sslPrivateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch the SSL private key: %w", err)
	}

	return certif, key, nil
}

func (s *ConfigService) setPEMPath(ctx context.Context, key ConfigKey, path string) error {
	rawCertif, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(rawCertif)
	if block == nil {
		return ErrInvalidPEMFormat
	}

	return s.storage.Save(ctx, key, path)
}

func (s *ConfigService) SetTrustedHosts(ctx context.Context, hosts []string) error {
	toSave := make([]string, len(hosts))

	for idx, host := range hosts {
		err := is.Host.Validate(host)
		if err != nil {
			return fmt.Errorf("invalid host %q: %w", host, err)
		}

		if ip := net.ParseIP(host); ip != nil {
			toSave[idx] = ip.String()
		} else {
			toSave[idx] = hosts[idx]
		}
	}

	err := s.storage.Save(ctx, hostsTrusted, strings.Join(hosts, ","))
	if err != nil {
		return fmt.Errorf("failed to save into the storage: %w", err)
	}

	return nil
}

func (s *ConfigService) GetTrustedHosts(ctx context.Context) ([]string, error) {
	rawRes, err := s.storage.Get(ctx, hostsTrusted)
	if errors.Is(err, errNotfound) {
		return nil, ErrNotInitialized
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get from the storage: %w", err)
	}

	return strings.Split(rawRes, ","), nil
}

func (s *ConfigService) SetHostName(ctx context.Context, input string) error {
	host, port, _ := net.SplitHostPort(input)
	if host == "" {
		host = input
	}

	// Host validates if a string is a valid IP (both v4 and v6) or a valid DNS name
	err := is.Host.Validate(host)
	if err != nil {
		return fmt.Errorf("invalid hostname: %w", err)
	}

	if port != "" {
		host = net.JoinHostPort(host, port)
	}

	err = s.storage.Save(ctx, hostName, host)
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *ConfigService) GetHostName(ctx context.Context) (string, error) {
	res, err := s.storage.Get(ctx, hostName)
	if errors.Is(err, errNotfound) {
		return "", ErrNotInitialized
	}

	if err != nil {
		return "", fmt.Errorf("failed to get from the storage: %w", err)
	}

	return res, nil
}

func (s *ConfigService) SetAddrs(ctx context.Context, hosts []string, port int) error {
	addrs := make([]string, len(hosts))

	portStr := strconv.Itoa(port)

	err := is.Port.Validate(portStr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPort, err)
	}

	for idx, host := range hosts {
		hostWithoutPort, _, _ := net.SplitHostPort(host)
		if hostWithoutPort != "" {
			return ErrMustNotHavePort
		}

		err := is.Host.Validate(host)
		if err != nil {
			return fmt.Errorf("invalid host %q: %w", host, err)
		}

		addrs[idx] = net.JoinHostPort(host, portStr)
	}

	err = s.storage.Save(ctx, httpAddrs, strings.Join(addrs, ","))
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *ConfigService) GetAddrs(ctx context.Context) ([]string, error) {
	res, err := s.storage.Get(ctx, httpAddrs)
	if errors.Is(err, errNotfound) {
		return nil, ErrNotInitialized
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get from the storage: %w", err)
	}

	return strings.Split(res, ","), nil
}
