package services

import (
	"context"
	"sort"
	"time"

	"analytics-service/internal/cache"
)

const readMaxPeriods = 12

type MetricsReadService struct {
	cache cache.Cache
}

func NewMetricsReadService(cache cache.Cache) *MetricsReadService {
	return &MetricsReadService{cache: cache}
}

func (s *MetricsReadService) Summary(ctx context.Context, from, to *time.Time) (cache.SummaryPayload, bool, error) {
	snapshot, found, err := s.cache.GetUserMetrics(ctx)
	if err != nil || !found {
		return cache.SummaryPayload{}, found, err
	}

	cohort := filterCohort(snapshot.Users, from, to)
	return summaryFromUsers(cohort, time.Now().UTC()), true, nil
}

func (s *MetricsReadService) Retention(
	ctx context.Context,
	planId int16,
	from, to *time.Time,
	maxPeriods int,
) (cache.RetentionPayload, bool, error) {
	snapshot, found, err := s.cache.GetUserMetrics(ctx)
	if err != nil || !found {
		return cache.RetentionPayload{}, found, err
	}

	cohort := filterCohort(snapshot.Users, from, to)
	return retentionFromUsers(cohort, planId, time.Now().UTC(), maxPeriods), true, nil
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
	for i := 0; i < maxPeriods; i++ {
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
