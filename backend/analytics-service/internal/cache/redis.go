package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	summaryKey   = "analytics:ltv:summary"
	retentionKey = "analytics:retention"
	userMetricsKey = "analytics:users:metrics"
)

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

type RetentionPeriodPayload struct {
	Period         int     `json:"period"`
	Eligible       int     `json:"eligible"`
	Renewed        int     `json:"renewed"`
	Rate           float64 `json:"rate"`
	CumulativeRate float64 `json:"cumulative_rate"`
}

type RetentionPayload struct {
	PlanId     int16                   `json:"plan_id"`
	DataAsOf   time.Time               `json:"data_as_of"`
	CohortSize int                     `json:"cohort_size"`
	Periods    []RetentionPeriodPayload `json:"periods"`
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

func (c *RedisCache) GetSummary(ctx context.Context) (SummaryPayload, bool, error) {
	raw, err := c.rdb.Get(ctx, summaryKey).Result()
	if isMiss(err) {
		return SummaryPayload{}, false, nil
	}
	if err != nil {
		return SummaryPayload{}, false, fmt.Errorf("cache: get summary: %w", err)
	}

	var summary SummaryPayload
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return SummaryPayload{}, false, fmt.Errorf("cache: unmarshal summary: %w", err)
	}

	return summary, true, nil
}

func (c *RedisCache) GetRetention(ctx context.Context) ([]RetentionPayload, bool, error) {
	raw, err := c.rdb.Get(ctx, retentionKey).Result()
	if isMiss(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("cache: get retention: %w", err)
	}

	var retentions []RetentionPayload
	if err := json.Unmarshal([]byte(raw), &retentions); err != nil {
		return nil, false, fmt.Errorf("cache: unmarshal retention: %w", err)
	}

	return retentions, true, nil
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

func (c *RedisCache) GetUserMetrics(ctx context.Context) (UserMetricsSnapshot, bool, error) {
	raw, err := c.rdb.Get(ctx, userMetricsKey).Result()
	if isMiss(err) {
		return UserMetricsSnapshot{}, false, nil
	}
	if err != nil {
		return UserMetricsSnapshot{}, false, fmt.Errorf("cache: get users metrics: %w", err)
	}

	var snapshot UserMetricsSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return UserMetricsSnapshot{}, false, fmt.Errorf("cache: unmarshal users metrics: %w", err)
	}

	return snapshot, true, nil
}

func isMiss(err error) bool {
	return err == redis.Nil
}
