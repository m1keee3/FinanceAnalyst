package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/cache"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, db int) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *Cache) GetStats(ctx context.Context) ([]models.SegmentStats, error) {
	const op = "cache.Redis.GetStats"

	var cursor uint64
	var result []models.SegmentStats

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, "segment_stats:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if len(keys) > 0 {
			values, err := c.client.MGet(ctx, keys...).Result()
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			for _, v := range values {
				if v == nil {
					continue
				}

				var stat models.SegmentStats
				if err := json.Unmarshal([]byte(v.(string)), &stat); err != nil {
					return nil, fmt.Errorf("%s: %w", op, err)
				}

				result = append(result, stat)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if len(result) == 0 {
		return nil, cache.ErrNotFound
	}

	return result, nil
}

func (c *Cache) SetSegmentStats(ctx context.Context, segStats models.SegmentStats, ttl time.Duration) error {
	const op = "cache.Redis.SetSegmentStats"

	data, err := json.Marshal(segStats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.client.Set(ctx, fmt.Sprintf("segment_stats:%s:%s", segStats.PatternType, segStats.Segment.Ticker), data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cache) Clear(ctx context.Context) error {
	const op = "cache.Redis.Clear"

	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, "segment_stats:*", 100).Result()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
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
