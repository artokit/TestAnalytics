package services

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"analytics-service/internal/cache"
)

type concurrentCache struct {
	mu       sync.Mutex
	snapshot cache.UserMetricsSnapshot
	computed map[string][]byte
	saves    map[string]int
}

func newConcurrentCache(snapshot cache.UserMetricsSnapshot) *concurrentCache {
	return &concurrentCache{
		snapshot: snapshot,
		computed: make(map[string][]byte),
		saves:    make(map[string]int),
	}
}

func (c *concurrentCache) SaveSummary(context.Context, cache.SummaryPayload) error { return nil }

func (c *concurrentCache) SaveRetention(context.Context, []cache.RetentionPayload) error { return nil }

func (c *concurrentCache) SaveUserMetrics(context.Context, cache.UserMetricsSnapshot) error {
	return nil
}

func (c *concurrentCache) GetUserMetrics(context.Context) (cache.UserMetricsSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshot, nil
}

func (c *concurrentCache) SaveComputed(_ context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.computed[key] = raw
	c.saves[key]++
	c.mu.Unlock()

	return nil
}

func (c *concurrentCache) GetComputed(_ context.Context, key string, dst any) error {
	c.mu.Lock()
	raw, ok := c.computed[key]
	c.mu.Unlock()

	if !ok {
		time.Sleep(time.Millisecond)
		return cache.ErrNotFound
	}

	return json.Unmarshal(raw, dst)
}

func (c *concurrentCache) savesFor(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saves[key]
}

func raceSnapshot(t *testing.T) cache.UserMetricsSnapshot {
	t.Helper()

	base := testBase()
	snapshot := cache.UserMetricsSnapshot{GeneratedAt: base.Add(24 * time.Hour)}

	for i := 0; i < 8; i++ {
		planId := int16(1 + i%2)
		months := 1
		if planId == 2 {
			months = 12
		}
		start := base.Add(time.Duration(i) * 24 * time.Hour)

		u := cache.UserMetrics{
			UserId:          fmt.Sprintf("user%d", i),
			Ltv:             int64(100 * (i + 1)),
			FirstPlanId:     planId,
			FirstPlanMonths: months,
			FirstSubDate:    start,
			Chain: []cache.ChainItem{
				{PlanId: planId, PeriodStart: start, PeriodEnd: start.Add(time.Duration(months) * testDay)},
			},
		}

		if i%4 == 1 {
			next := start.Add(time.Duration(months)*testDay + 24*time.Hour)
			u.Chain = append(u.Chain, cache.ChainItem{
				PlanId:      planId,
				PeriodStart: next,
				PeriodEnd:   next.Add(time.Duration(months) * testDay),
			})
		}

		snapshot.Users = append(snapshot.Users, u)
	}

	return snapshot
}

func TestReadServiceConcurrentSameKeys(t *testing.T) {
	base := testBase()
	end := base.Add(400 * 24 * time.Hour)

	snapshot := raceSnapshot(t)
	fc := newConcurrentCache(snapshot)
	svc := NewMetricsReadService(fc, time.Hour, nil)

	const workers = 16

	summaries := make([]cache.SummaryPayload, workers)
	retentions := make([]cache.RetentionPayload, workers)
	byPlans := make([]cache.LtvByPlanPayload, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx := context.Background()

			summary, err := svc.Summary(ctx, &base, &end)
			if err != nil {
				t.Errorf("summary: %v", err)
				return
			}
			summaries[i] = summary

			retention, err := svc.Retention(ctx, 1, &base, &end, 12)
			if err != nil {
				t.Errorf("retention: %v", err)
				return
			}
			retentions[i] = retention

			byPlan, err := svc.LtvByPlan(ctx, &base, &end)
			if err != nil {
				t.Errorf("ltv by plan: %v", err)
				return
			}
			byPlans[i] = byPlan
		}(i)
	}
	wg.Wait()

	for i := 1; i < workers; i++ {
		if summaries[i] != summaries[0] {
			t.Errorf("summary mismatch at worker %d: %+v vs %+v", i, summaries[i], summaries[0])
		}
		if !reflect.DeepEqual(retentions[i], retentions[0]) {
			t.Errorf("retention mismatch at worker %d: %+v vs %+v", i, retentions[i], retentions[0])
		}
		if !reflect.DeepEqual(byPlans[i], byPlans[0]) {
			t.Errorf("by-plan mismatch at worker %d: %+v vs %+v", i, byPlans[i], byPlans[0])
		}
	}

	keys := []string{
		summaryKey(&base, &end, snapshot.GeneratedAt),
		retentionKey(1, 12, &base, &end, snapshot.GeneratedAt),
		ltvByPlanKey(&base, &end, snapshot.GeneratedAt),
	}
	for _, key := range keys {
		if got := fc.savesFor(key); got != 1 {
			t.Errorf("computed saves for %s = %d, want exactly 1", key, got)
		}
	}
}

func TestReadServiceConcurrentMixedFilters(t *testing.T) {
	base := testBase()
	mid := base.Add(100 * 24 * time.Hour)

	snapshot := raceSnapshot(t)
	fc := newConcurrentCache(snapshot)
	svc := NewMetricsReadService(fc, time.Hour, nil)

	ranges := []struct {
		from, to *time.Time
		name     string
	}{
		{nil, nil, "all"},
		{&base, nil, "from base"},
		{nil, &mid, "to mid"},
		{&base, &mid, "base..mid"},
	}

	const workers = 24
	summaries := make([][]cache.SummaryPayload, len(ranges))
	byPlans := make([][]cache.LtvByPlanPayload, len(ranges))
	groups := make([]int, workers)

	var collect sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		group := i % len(ranges)
		groups[i] = group

		wg.Add(1)
		go func(i, group int) {
			defer wg.Done()
			ctx := context.Background()
			r := ranges[group]

			summary, err := svc.Summary(ctx, r.from, r.to)
			if err != nil {
				t.Errorf("summary: %v", err)
				return
			}

			byPlan, err := svc.LtvByPlan(ctx, r.from, r.to)
			if err != nil {
				t.Errorf("ltv by plan: %v", err)
				return
			}

			collect.Lock()
			summaries[group] = append(summaries[group], summary)
			byPlans[group] = append(byPlans[group], byPlan)
			collect.Unlock()
		}(i, group)
	}
	wg.Wait()

	for gi, r := range ranges {
		for i := 1; i < len(summaries[gi]); i++ {
			if summaries[gi][i] != summaries[gi][0] {
				t.Errorf("filter %s: summary mismatch: %+v vs %+v", r.name, summaries[gi][i], summaries[gi][0])
			}
			if !reflect.DeepEqual(byPlans[gi][i], byPlans[gi][0]) {
				t.Errorf("filter %s: by-plan mismatch: %+v vs %+v", r.name, byPlans[gi][i], byPlans[gi][0])
			}
		}

		keys := []string{
			summaryKey(r.from, r.to, snapshot.GeneratedAt),
			ltvByPlanKey(r.from, r.to, snapshot.GeneratedAt),
		}
		for _, key := range keys {
			if got := fc.savesFor(key); got != 1 {
				t.Errorf("filter %s: computed saves for %s = %d, want exactly 1", r.name, key, got)
			}
		}
	}
}
