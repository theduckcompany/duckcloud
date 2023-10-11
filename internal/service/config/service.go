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

func (s *ConfigService) Get(ctx context.Context, key ConfigKey) (string, error) {
	return s.storage.Get(ctx, key)
}

func (s *ConfigService) EnableTLS(ctx context.Context) error {
	return s.storage.Save(ctx, TLSEnabled, "true")
}

func (s *ConfigService) EnableDevMode(ctx context.Context) error {
	return s.storage.Save(ctx, DevModeEnabled, "true")
}

func (s *ConfigService) IsDevModeEnabled(ctx context.Context) (bool, error) {
	return s.checkBool(ctx, DevModeEnabled)
}

func (s *ConfigService) DisableTLS(ctx context.Context) error {
	return s.storage.Save(ctx, TLSEnabled, "false")
}

func (s *ConfigService) IsTLSEnabled(ctx context.Context) (bool, error) {
	return s.checkBool(ctx, TLSEnabled)
}

func (s *ConfigService) checkBool(ctx context.Context, key ConfigKey) (bool, error) {
	enabled, err := s.storage.Get(ctx, key)
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
		return errors.New("SSL must be enabled")
	}

	err = s.setPEMPath(ctx, SSLCertificatePath, certifPath)
	if err != nil {
		return fmt.Errorf("failed to save the certificate: %w", err)
	}

	err = s.setPEMPath(ctx, SSLPrivateKeyPath, privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to save the private key: %w", err)
	}

	return nil
}

func (s *ConfigService) setPEMPath(ctx context.Context, key ConfigKey, path string) error {
	rawCertif, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(rawCertif)
	if block == nil {
		return errors.New("invalid PEM format")
	}

	return s.storage.Save(ctx, key, path)
}

func (s *ConfigService) setPath(ctx context.Context, key ConfigKey, path string, expectDir bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() != expectDir {
		return fmt.Errorf("%q must be a directory", path)
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

	err := s.storage.Save(ctx, HostsTrusted, strings.Join(hosts, ","))
	if err != nil {
		return fmt.Errorf("failed to save into the storage: %w", err)
	}

	return nil
}

func (s *ConfigService) GetTrustedHosts(ctx context.Context) ([]string, error) {
	rawRes, err := s.storage.Get(ctx, HostsTrusted)
	if err != nil {
		return nil, fmt.Errorf("failed to get from the storage: %w", err)
	}

	return strings.Split(rawRes, ","), nil
}

func (s *ConfigService) SetHostName(ctx context.Context, hostName string) error {
	host, port, _ := net.SplitHostPort(hostName)
	if host == "" {
		host = hostName
	}

	// Host validates if a string is a valid IP (both v4 and v6) or a valid DNS name
	err := is.Host.Validate(host)
	if err != nil {
		return fmt.Errorf("invalid hostname: %w", err)
	}

	if port != "" {
		host = net.JoinHostPort(host, port)
	}

	err = s.storage.Save(ctx, HostName, host)
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *ConfigService) SetAddrs(ctx context.Context, hosts []string, port int) error {
	addrs := make([]string, len(hosts))

	portStr := strconv.Itoa(port)

	err := is.Port.Validate(portStr)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", port, err)
	}

	for idx, host := range hosts {
		err := is.Host.Validate(host)
		if err != nil {
			return fmt.Errorf("invalid host %q: %w", host, err)
		}

		addrs[idx] = net.JoinHostPort(host, portStr)
	}

	err = s.storage.Save(ctx, HTTPAddrs, strings.Join(addrs, ","))
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}
