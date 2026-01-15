package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr string, password string, db int) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (r Cache) GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error) {
	const op = "cache.redis.GetScan"

	data, err := r.client.Get(ctx, fmt.Sprintf("scan:%s", hash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s: %w", op, cache.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var segments []models.ChartSegment
	if err := json.Unmarshal(data, &segments); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return segments, nil
}

func (r Cache) SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error {
	const op = "cache.redis.SetScan"

	data, err := json.Marshal(segments)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.client.Set(ctx, fmt.Sprintf("scan:%s", hash), data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
