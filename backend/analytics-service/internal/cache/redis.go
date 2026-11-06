package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("cache: snapshot not found")

const (
	summaryKey     = "analytics:ltv:summary"
	retentionKey   = "analytics:retention"
	userMetricsKey = "analytics:users:metrics:v3"
	computedPrefix = "analytics:computed:"
)

func computedKey(key string) string {
	return computedPrefix + key
}

type Cache interface {
	SaveSummary(ctx context.Context, summary SummaryPayload) error
	SaveRetention(ctx context.Context, retentions []RetentionPayload) error
	SaveUserMetrics(ctx context.Context, snapshot UserMetricsSnapshot) error
	GetUserMetrics(ctx context.Context) (UserMetricsSnapshot, error)
	SaveComputed(ctx context.Context, key string, value any) error
	GetComputed(ctx context.Context, key string, dst any) error
}

type RedisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewCache(rdb *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		rdb: rdb,
		ttl: ttl,
	}
}

type SummaryPayload struct {
	DataAsOf     time.Time `json:"data_as_of"`
	AvgLtv       int64     `json:"avg_ltv"`
	MedianLtv    int64     `json:"median_ltv"`
	TotalRevenue int64     `json:"total_revenue"`
	Users        int       `json:"users"`
}

func (c *RedisCache) SaveSummary(ctx context.Context, summary SummaryPayload) error {
	raw, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("cache: marshal summary: %w", err)
	}

	if err := c.rdb.Set(ctx, summaryKey, raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache: set summary: %w", err)
	}

	return nil
}

func (c *RedisCache) SaveRetention(ctx context.Context, retentions []RetentionPayload) error {
	raw, err := json.Marshal(retentions)
	if err != nil {
		return fmt.Errorf("cache: marshal retention: %w", err)
	}

	if err := c.rdb.Set(ctx, retentionKey, raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache: set retention: %w", err)
	}

	return nil
}

func (c *RedisCache) GetSummary(ctx context.Context) (SummaryPayload, error) {
	raw, err := c.rdb.Get(ctx, summaryKey).Result()
	if err == redis.Nil {
		return SummaryPayload{}, ErrNotFound
	}
	if err != nil {
		return SummaryPayload{}, fmt.Errorf("cache: get summary: %w", err)
	}

	var summary SummaryPayload
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return SummaryPayload{}, fmt.Errorf("cache: unmarshal summary: %w", err)
	}

	return summary, nil
}

func (c *RedisCache) GetRetention(ctx context.Context) ([]RetentionPayload, error) {
	raw, err := c.rdb.Get(ctx, retentionKey).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get retention: %w", err)
	}

	var retentions []RetentionPayload
	if err := json.Unmarshal([]byte(raw), &retentions); err != nil {
		return nil, fmt.Errorf("cache: unmarshal retention: %w", err)
	}

	return retentions, nil
}

func (c *RedisCache) SaveUserMetrics(ctx context.Context, snapshot UserMetricsSnapshot) error {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("cache: marshal users metrics: %w", err)
	}

	if err := c.rdb.Set(ctx, userMetricsKey, raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache: set users metrics: %w", err)
	}

	return nil
}

func (c *RedisCache) GetUserMetrics(ctx context.Context) (UserMetricsSnapshot, error) {
	raw, err := c.rdb.Get(ctx, userMetricsKey).Result()
	if err == redis.Nil {
		return UserMetricsSnapshot{}, ErrNotFound
	}
	if err != nil {
		return UserMetricsSnapshot{}, fmt.Errorf("cache: get users metrics: %w", err)
	}

	var snapshot UserMetricsSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return UserMetricsSnapshot{}, fmt.Errorf("cache: unmarshal users metrics: %w", err)
	}

	return snapshot, nil
}

func (c *RedisCache) SaveComputed(ctx context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal %s: %w", key, err)
	}

	if err := c.rdb.Set(ctx, computedKey(key), raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache: set %s: %w", key, err)
	}

	return nil
}

func (c *RedisCache) GetComputed(ctx context.Context, key string, dst any) error {
	raw, err := c.rdb.Get(ctx, computedKey(key)).Result()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cache: get %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return fmt.Errorf("cache: unmarshal %s: %w", key, err)
	}

	return nil
}
