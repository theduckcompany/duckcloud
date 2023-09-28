package config

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

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

func (s *ConfigService) Bootstrap(ctx context.Context, cmd BootstrapCmd) error {
	res, err := s.storage.Get(ctx, HostName)
	if err != nil {
		return fmt.Errorf("failed to get %q: %w", HostName, err)
	}

	if res != "" {
		return ErrAlreadyBootstraped
	}

	addrError := new(net.AddrError)

	host, port, err := net.SplitHostPort(cmd.HostName)
	if errors.As(err, &addrError) {
	}
	switch {
	case err == nil:
		if err := is.Port.Validate(port); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidHostname, err)
		}
		break
	case errors.As(err, &addrError) && addrError.Err == "missing port in address":
		// there is juste the host so use the proper validation
	default:
		return fmt.Errorf("%w: %w", ErrInvalidHostname, err)
	}

	if err := is.Host.Validate(host); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHostname, err)
	}

	fullHost := net.JoinHostPort(host, port)

	err = s.storage.Save(ctx, HostName, fullHost)
	if err != nil {
		return fmt.Errorf("failed to save %q: %w", HostName, err)
	}

	err = s.storage.Save(ctx, Port, strconv.Itoa(cmd.Port))
	if err != nil {
		return fmt.Errorf("failed to save %q: %w", Port, err)
	}

	return nil
}
