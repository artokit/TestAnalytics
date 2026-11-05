package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"analytics-service/internal/cache"

	"golang.org/x/sync/singleflight"
)

var ErrSnapshotNotReady = errors.New("metrics: data snapshot is not ready")

type MetricsReadService struct {
	cache    cache.Cache
	ttl      time.Duration
	inflight singleflight.Group
}

func NewMetricsReadService(cache cache.Cache, ttl time.Duration) *MetricsReadService {
	return &MetricsReadService{
		cache: cache,
		ttl:   ttl,
	}
}

func (s *MetricsReadService) Summary(ctx context.Context, from, to *time.Time) (cache.SummaryPayload, error) {
	snapshot, err := s.cache.GetUserMetrics(ctx)
	if errors.Is(err, cache.ErrNotFound) {
		return cache.SummaryPayload{}, ErrSnapshotNotReady
	}
	if err != nil {
		return cache.SummaryPayload{}, err
	}

	key := summaryKey(from, to, snapshot.GeneratedAt)
	value, err, _ := s.inflight.Do(key, func() (any, error) {
		var payload cache.SummaryPayload
		if err := s.cache.GetComputed(ctx, key, &payload); err == nil {
			return payload, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		}

		summary := summaryFromUsers(filterCohort(snapshot.Users, from, to), snapshot.GeneratedAt)
		return summary, s.cache.SaveComputed(ctx, key, summary)
	})
	if err != nil {
		return cache.SummaryPayload{}, err
	}

	return value.(cache.SummaryPayload), nil
}

func (s *MetricsReadService) Retention(
	ctx context.Context,
	planId int16,
	from, to *time.Time,
	maxPeriods int,
) (cache.RetentionPayload, error) {
	snapshot, err := s.cache.GetUserMetrics(ctx)
	if errors.Is(err, cache.ErrNotFound) {
		return cache.RetentionPayload{}, ErrSnapshotNotReady
	}
	if err != nil {
		return cache.RetentionPayload{}, err
	}

	key := retentionKey(planId, maxPeriods, from, to, snapshot.GeneratedAt)
	value, err, _ := s.inflight.Do(key, func() (any, error) {
		var payload cache.RetentionPayload
		if err := s.cache.GetComputed(ctx, key, &payload); err == nil {
			return payload, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		}

		retention := retentionFromUsers(filterCohort(snapshot.Users, from, to), planId, snapshot.GeneratedAt, maxPeriods)
		return retention, s.cache.SaveComputed(ctx, key, retention)
	})
	if err != nil {
		return cache.RetentionPayload{}, err
	}

	return value.(cache.RetentionPayload), nil
}

func summaryKey(from, to *time.Time, generatedAt time.Time) string {
	key := "summary:generated_at=" + generatedAt.Format(time.RFC3339Nano)
	if from != nil {
		key += ":from=" + from.Format(time.RFC3339Nano)
	}
	if to != nil {
		key += ":to=" + to.Format(time.RFC3339Nano)
	}
	return key
}

func retentionKey(planId int16, maxPeriods int, from, to *time.Time, generatedAt time.Time) string {
	key := fmt.Sprintf("retention:plan=%d:periods=%d:generated_at=%s", planId, maxPeriods, generatedAt.Format(time.RFC3339Nano))
	if from != nil {
		key += ":from=" + from.Format(time.RFC3339Nano)
	}
	if to != nil {
		key += ":to=" + to.Format(time.RFC3339Nano)
	}
	return key
}

func filterCohort(users []cache.UserMetrics, from, to *time.Time) []cache.UserMetrics {
	out := make([]cache.UserMetrics, 0, len(users))
	for _, u := range users {
		if inRange(u.FirstSubDate, from, to) {
			out = append(out, u)
		}
	}
	return out
}

func inRange(t time.Time, from, to *time.Time) bool {
	if from != nil && t.Before(*from) {
		return false
	}
	if to != nil && t.After(*to) {
		return false
	}
	return true
}

func summaryFromUsers(users []cache.UserMetrics, now time.Time) cache.SummaryPayload {
	ltvs := make([]int64, 0, len(users))
	var total int64
	for _, u := range users {
		ltvs = append(ltvs, u.Ltv)
		total += u.Ltv
	}
	sort.Slice(ltvs, func(i, j int) bool { return ltvs[i] < ltvs[j] })

	avg := int64(0)
	if len(users) > 0 {
		avg = total / int64(len(users))
	}

	return cache.SummaryPayload{
		DataAsOf:     now,
		AvgLtv:       avg,
		MedianLtv:    medianCohort(len(users), ltvs),
		TotalRevenue: total,
		Users:        len(users),
	}
}

func retentionFromUsers(
	users []cache.UserMetrics,
	planId int16,
	now time.Time,
	maxPeriods int,
) cache.RetentionPayload {
	eligible := make([]int, maxPeriods)
	renewed := make([]int, maxPeriods)
	cohortSize := 0

	for _, u := range users {
		if u.FirstPlanId != planId || len(u.Chain) == 0 {
			continue
		}
		cohortSize++

		for i, row := range u.Chain {
			period := i + 1
			if period > maxPeriods {
				break
			}
			if row.PeriodEnd.After(now) {
				break
			}
			eligible[period-1]++

			if i+1 < len(u.Chain) && !u.Chain[i+1].PeriodStart.Before(row.PeriodEnd) {
				renewed[period-1]++
			}
		}
	}

	periods := make([]cache.RetentionPeriodPayload, 0, maxPeriods)
	for i := range maxPeriods {
		if eligible[i] == 0 && renewed[i] == 0 {
			break
		}

		rate := float64(0)
		if eligible[i] > 0 {
			rate = float64(renewed[i]) / float64(eligible[i])
		}
		cumulative := float64(0)
		if cohortSize > 0 {
			cumulative = float64(renewed[i]) / float64(cohortSize)
		}

		periods = append(periods, cache.RetentionPeriodPayload{
			Period:         i + 1,
			Eligible:       eligible[i],
			Renewed:        renewed[i],
			Rate:           rate,
			CumulativeRate: cumulative,
		})
	}

	return cache.RetentionPayload{
		PlanId:     planId,
		DataAsOf:   now,
		CohortSize: cohortSize,
		Periods:    periods,
	}
}
