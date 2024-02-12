package stats

import (
	"context"
	"fmt"
	"strconv"
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, key statsKey, value any) error
	Get(ctx context.Context, key statsKey, val any) error
}

type ConfigService struct {
	storage Storage
}

func NewService(storage Storage) *ConfigService {
	return &ConfigService{storage}
}

func (s *ConfigService) SetTotalSize(ctx context.Context, totalSize uint64) error {
	err := s.storage.Save(ctx, totalSizeKey, strconv.FormatUint(totalSize, 10))
	if err != nil {
		return fmt.Errorf("failed to Save: %w", err)
	}

	return nil
}

func (s *ConfigService) GetTotalSize(ctx context.Context) (uint64, error) {
	var resStr string

	err := s.storage.Get(ctx, totalSizeKey, &resStr)
	if err != nil {
		return 0, fmt.Errorf("failed to Get: %w", err)
	}

	return strconv.ParseUint(resStr, 10, 64)
}
