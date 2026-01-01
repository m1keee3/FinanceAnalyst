package redis

import (
	"context"
	"errors"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache"
)

type Cache struct{}

func NewCache() *Cache {
	return &Cache{}
}

func (r Cache) GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error) {
	//TODO implement me
	return nil, cache.ErrNotFound
}

func (r Cache) SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error {
	//TODO implement me
	return errors.New("not implemented")
}
