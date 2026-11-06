package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"analytics-service/internal/cache"
	"analytics-service/internal/domain"
)

const maxPeriods = 12

type MetricsService struct {
	cache cache.Cache
}

func NewMetricsService(cache cache.Cache) *MetricsService {
	return &MetricsService{cache: cache}
}

func (s *MetricsService) Recalculate(
	ctx context.Context,
	users map[string]domain.User,
	payments map[string]domain.Payment,
	subs map[string]domain.Subscription,
	plans map[int]domain.Plan,
	now time.Time,
) error {
	if err := s.cache.SaveSummary(ctx, buildSummary(users, payments, now)); err != nil {
		return fmt.Errorf("metrics: save summary: %w", err)
	}

	if err := s.cache.SaveRetention(ctx, buildRetention(subs, now)); err != nil {
		return fmt.Errorf("metrics: save retention: %w", err)
	}

	if err := s.cache.SaveUserMetrics(ctx, buildUserMetrics(users, payments, subs, plans, now)); err != nil {
		return fmt.Errorf("metrics: save users metrics: %w", err)
	}

	return nil
}

func buildUserMetrics(
	users map[string]domain.User,
	payments map[string]domain.Payment,
	subs map[string]domain.Subscription,
	plans map[int]domain.Plan,
	now time.Time,
) cache.UserMetricsSnapshot {
	ltvByUser := make(map[string]int64, len(users))
	paymentsByUser := make(map[string]int, len(users))
	for _, p := range payments {
		if _, ok := users[p.UserId]; !ok {
			continue
		}
		paymentsByUser[p.UserId]++
		if p.Status == "succeeded" {
			ltvByUser[p.UserId] += p.Amount
		}
	}

	chains := make(map[string][]domain.Subscription)
	for _, sub := range subs {
		chains[sub.UserId] = append(chains[sub.UserId], sub)
	}

	planMonths := make(map[int16]int, len(plans))
	for id, plan := range plans {
		planMonths[int16(id)] = plan.PeriodMonths
	}

	payload := cache.UserMetricsSnapshot{
		GeneratedAt: now,
		Users:       make([]cache.UserMetrics, 0, len(users)),
	}

	for userId := range users {
		user := cache.UserMetrics{
			UserId:        userId,
			Ltv:           ltvByUser[userId],
			TelegramId:    users[userId].TelegramId,
			PaymentsCount: paymentsByUser[userId],
		}

		chain := chains[userId]
		if len(chain) > 0 {
			sort.Slice(chain, func(i, j int) bool { return chain[i].PeriodStart.Before(chain[j].PeriodStart) })
			user.FirstPlanId = chain[0].PlanId
			user.FirstSubDate = chain[0].PeriodStart
			user.FirstPlanMonths = firstPlanMonths(planMonths, chain[0])
			user.SubStatus = chain[len(chain)-1].Status
			for _, row := range chain {
				user.Chain = append(user.Chain, cache.ChainItem{
					PlanId:      row.PlanId,
					PeriodStart: row.PeriodStart,
					PeriodEnd:   row.PeriodEnd,
				})
			}
		}

		payload.Users = append(payload.Users, user)
	}

	sort.Slice(payload.Users, func(i, j int) bool { return payload.Users[i].UserId < payload.Users[j].UserId })

	return payload
}

func firstPlanMonths(planMonths map[int16]int, first domain.Subscription) int {
	if m, ok := planMonths[first.PlanId]; ok {
		return m
	}

	months := int(first.PeriodEnd.Sub(first.PeriodStart) / (30 * 24 * time.Hour))
	if months < 1 {
		return 1
	}
	return months
}

func buildSummary(
	users map[string]domain.User,
	payments map[string]domain.Payment,
	now time.Time,
) cache.SummaryPayload {
	ltvByUser := make(map[string]int64, len(users))

	for _, p := range payments {
		switch p.Status {
		case "succeeded":
			ltvByUser[p.UserId] += p.Amount
		case "refunded":
			ltvByUser[p.UserId] -= p.Amount
		}
	}

	ltvs := make([]int64, 0, len(ltvByUser))
	var total int64
	for _, ltv := range ltvByUser {
		ltvs = append(ltvs, ltv)
		total += ltv
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

func medianCohort(cohortSize int, sorted []int64) int64 {
	if cohortSize == 0 {
		return 0
	}

	at := func(k int) int64 {
		if k < cohortSize-len(sorted) {
			return 0
		}
		return sorted[k-(cohortSize-len(sorted))]
	}

	lo, hi := at((cohortSize-1)/2), at(cohortSize/2)
	return (lo + hi) / 2
}

func buildRetention(
	subs map[string]domain.Subscription,
	now time.Time,
) []cache.RetentionPayload {
	type retentionAgg struct {
		cohortSize int
		eligible   [maxPeriods]int
		renewed    [maxPeriods]int
	}

	chains := make(map[string][]domain.Subscription, len(subs))
	firstPlan := make(map[string]int16, len(subs))
	for _, sub := range subs {
		chains[sub.UserId] = append(chains[sub.UserId], sub)
	}
	for userId, chain := range chains {
		sort.Slice(chain, func(i, j int) bool { return chain[i].PeriodStart.Before(chain[j].PeriodStart) })
		firstPlan[userId] = chain[0].PlanId
	}

	aggs := make(map[int16]*retentionAgg)
	for userId, chain := range chains {
		plan := firstPlan[userId]
		agg, ok := aggs[plan]
		if !ok {
			agg = &retentionAgg{}
			aggs[plan] = agg
		}
		agg.cohortSize++

		for i, row := range chain {
			period := i + 1
			if period > maxPeriods {
				break
			}
			if row.PeriodEnd.After(now) {
				break
			}
			agg.eligible[period-1]++

			if i+1 < len(chain) && !chain[i+1].PeriodStart.Before(row.PeriodEnd) {
				agg.renewed[period-1]++
			}
		}
	}

	payloads := make([]cache.RetentionPayload, 0, len(aggs))
	for planId, agg := range aggs {
		periods := make([]cache.RetentionPeriodPayload, 0, maxPeriods)
		for i := 0; i < maxPeriods; i++ {
			if agg.eligible[i] == 0 && agg.renewed[i] == 0 {
				break
			}

			rate := float64(0)
			if agg.eligible[i] > 0 {
				rate = float64(agg.renewed[i]) / float64(agg.eligible[i])
			}
			cumulative := float64(0)
			if agg.cohortSize > 0 {
				cumulative = float64(agg.renewed[i]) / float64(agg.cohortSize)
			}

			periods = append(periods, cache.RetentionPeriodPayload{
				Period:         i + 1,
				Eligible:       agg.eligible[i],
				Renewed:        agg.renewed[i],
				Rate:           rate,
				CumulativeRate: cumulative,
			})
		}

		payloads = append(payloads, cache.RetentionPayload{
			PlanId:     planId,
			DataAsOf:   now,
			CohortSize: agg.cohortSize,
			Periods:    periods,
		})
	}

	sort.Slice(payloads, func(i, j int) bool { return payloads[i].PlanId < payloads[j].PlanId })

	return payloads
}
