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

func (c *Cache) GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error) {
	const op = "cache.Redis.GetScan"

	data, err := c.client.Get(ctx, fmt.Sprintf("scan:%s", hash)).Bytes()
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

func (c *Cache) SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error {
	const op = "cache.Redis.SetScan"

	data, err := json.Marshal(segments)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.client.Set(ctx, fmt.Sprintf("scan:%s", hash), data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Cache) GetStats(ctx context.Context, hash string) (*models.ScanStats, error) {
	const op = "cache.Redis.GetStats"

	data, err := c.client.Get(ctx, fmt.Sprintf("stats:%s", hash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s: %w", op, cache.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var stats models.ScanStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &stats, nil
}

func (c *Cache) SetStats(ctx context.Context, hash string, stats *models.ScanStats, ttl time.Duration) error {
	const op = "cache.Redis.SetScan"

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.client.Set(ctx, fmt.Sprintf("stats:%s", hash), data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Cache) Close() error {
	const op = "cache.Redis.Close"

	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
