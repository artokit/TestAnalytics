package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	"analytics-service/internal/cache"
	"analytics-service/internal/domain"
)

const testDay = 30 * 24 * time.Hour

func testBase() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

func testPlans() map[int]domain.Plan {
	return map[int]domain.Plan{
		1: {Id: 1, Name: "monthly", PeriodMonths: 1, Price: 500, Currency: "RUB"},
		2: {Id: 2, Name: "annual", PeriodMonths: 12, Price: 4800, Currency: "RUB"},
	}
}

func TestBuildUserMetricsLtv(t *testing.T) {
	base := testBase()
	users := map[string]domain.User{"p": {}, "q": {}, "r": {}, "s": {}}
	payments := map[string]domain.Payment{
		"p1": {Id: "p1", UserId: "p", Amount: 500, Status: "succeeded"},
		"p2": {Id: "p2", UserId: "p", Amount: 300, Status: "refunded"},
		"p3": {Id: "p3", UserId: "p", Amount: 700, Status: "failed"},
		"q1": {Id: "q1", UserId: "q", Amount: 100, Status: "refunded"},
	}
	subs := map[string]domain.Subscription{
		"s1": {Id: "s1", UserId: "p", PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
		"s2": {Id: "s2", UserId: "p", PlanId: 2, PeriodStart: base.Add(testDay), PeriodEnd: base.Add(13 * testDay)},
		"s3": {Id: "s3", UserId: "s", PlanId: 2, PeriodStart: base, PeriodEnd: base.Add(12 * testDay)},
	}

	snapshot := buildUserMetrics(users, payments, subs, testPlans(), base)

	got := make(map[string]cache.UserMetrics, len(snapshot.Users))
	for _, u := range snapshot.Users {
		got[u.UserId] = u
	}

	if u := got["p"]; u.Ltv != 500 {
		t.Errorf("Ltv = %d, want 500", u.Ltv)
	}
	if u := got["q"]; u.Ltv != 0 {
		t.Errorf("refunded-only user Ltv = %d, want 0", u.Ltv)
	}
	if u := got["r"]; u.Ltv != 0 {
		t.Errorf("no-payment user Ltv = %d, want 0", u.Ltv)
	}

	u := got["p"]
	if u.FirstPlanId != 1 {
		t.Errorf("FirstPlanId = %d, want 1", u.FirstPlanId)
	}
	if u.FirstPlanMonths != 1 {
		t.Errorf("FirstPlanMonths = %d, want 1", u.FirstPlanMonths)
	}
	if len(u.Chain) != 2 || u.Chain[0].PlanId != 1 || u.Chain[1].PlanId != 2 {
		t.Errorf("Chain = %+v, want monthly then annual", u.Chain)
	}

	if s := got["s"]; s.FirstPlanMonths != 12 {
		t.Errorf("annual subscriber FirstPlanMonths = %d, want 12", s.FirstPlanMonths)
	}
}

func TestBuildUserMetricsPlanMonthsFallback(t *testing.T) {
	base := testBase()
	users := map[string]domain.User{"f1": {}, "f2": {}}
	subs := map[string]domain.Subscription{
		"f1": {Id: "f1", UserId: "f1", PlanId: 9, PeriodStart: base, PeriodEnd: base.Add(10 * testDay)},
		"f2": {Id: "f2", UserId: "f2", PlanId: 9, PeriodStart: base, PeriodEnd: base.Add(10 * 24 * time.Hour)},
	}

	snapshot := buildUserMetrics(users, nil, subs, map[int]domain.Plan{}, base)

	got := make(map[string]cache.UserMetrics, len(snapshot.Users))
	for _, u := range snapshot.Users {
		got[u.UserId] = u
	}
	if got["f1"].FirstPlanMonths != 10 {
		t.Errorf("FirstPlanMonths = %d, want 10 from period duration", got["f1"].FirstPlanMonths)
	}
	if got["f2"].FirstPlanMonths != 1 {
		t.Errorf("FirstPlanMonths = %d, want 1 for too short period", got["f2"].FirstPlanMonths)
	}
}

func TestMedianCohort(t *testing.T) {
	cases := []struct {
		name       string
		cohortSize int
		sorted     []int64
		want       int64
	}{
		{"empty", 0, []int64{}, 0},
		{"single", 1, []int64{10}, 10},
		{"odd padded", 3, []int64{10, 20}, 10},
		{"even full", 4, []int64{0, 10, 20, 30}, 15},
		{"zeros dominate", 5, []int64{10, 20}, 0},
		{"even padded", 4, []int64{10, 20}, 5},
		{"pair", 2, []int64{10, 30}, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := medianCohort(tc.cohortSize, tc.sorted); got != tc.want {
				t.Errorf("medianCohort(%d, %v) = %d, want %d", tc.cohortSize, tc.sorted, got, tc.want)
			}
		})
	}
}

func TestSummaryFromUsers(t *testing.T) {
	base := testBase()
	now := base.Add(24 * time.Hour)
	users := []cache.UserMetrics{
		{UserId: "a", Ltv: 1000, FirstPlanId: 1, FirstSubDate: base},
		{UserId: "b", Ltv: 0, FirstPlanId: 1, FirstSubDate: base},
		{UserId: "c", Ltv: 250, FirstPlanId: 1, FirstSubDate: base},
	}

	summary := summaryFromUsers(users, now)

	want := cache.SummaryPayload{
		DataAsOf:     now,
		AvgLtv:       416,
		MedianLtv:    250,
		TotalRevenue: 1250,
		Users:        3,
	}
	if !reflect.DeepEqual(summary, want) {
		t.Errorf("summary = %+v, want %+v", summary, want)
	}
}

func TestFilterCohort(t *testing.T) {
	base := testBase()
	users := []cache.UserMetrics{
		{UserId: "early", FirstSubDate: base},
		{UserId: "mid", FirstSubDate: base.Add(50 * 24 * time.Hour)},
		{UserId: "late", FirstSubDate: base.Add(600 * 24 * time.Hour)},
		{UserId: "never"},
	}

	filter, err := time.Parse(time.RFC3339, "2025-02-20T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	got := filterCohort(users, &base, &filter)
	if len(got) != 2 || got[0].UserId != "early" || got[1].UserId != "mid" {
		t.Errorf("filtered = %+v, want early and mid inclusive bounds", got)
	}
}

func retentionFixture() []cache.UserMetrics {
	base := testBase()
	return []cache.UserMetrics{
		{
			UserId:          "churn1",
			FirstPlanId:     1,
			FirstPlanMonths: 1,
			FirstSubDate:    base,
			Chain: []cache.ChainItem{
				{PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
			},
		},
		{
			UserId:          "renew2",
			FirstPlanId:     1,
			FirstPlanMonths: 1,
			FirstSubDate:    base,
			Chain: []cache.ChainItem{
				{PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
				{PlanId: 1, PeriodStart: base.Add(testDay + 24*time.Hour), PeriodEnd: base.Add(2*testDay + 24*time.Hour)},
			},
		},
		{
			UserId:          "switch",
			FirstPlanId:     1,
			FirstPlanMonths: 1,
			FirstSubDate:    base,
			Chain: []cache.ChainItem{
				{PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
				{PlanId: 2, PeriodStart: base.Add(testDay + 3*24*time.Hour), PeriodEnd: base.Add(testDay + 3*24*time.Hour + 12*testDay)},
			},
		},
		{
			UserId:          "resurrect",
			FirstPlanId:     1,
			FirstPlanMonths: 1,
			FirstSubDate:    base,
			Chain: []cache.ChainItem{
				{PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
				{PlanId: 1, PeriodStart: base.Add(200 * 24 * time.Hour), PeriodEnd: base.Add(300 * 24 * time.Hour)},
			},
		},
	}
}

func TestRetentionFromUsersMonthly(t *testing.T) {
	users := retentionFixture()
	now := testBase().Add(400 * 24 * time.Hour)

	got := retentionFromUsers(users, 1, now, 12)

	if got.CohortSize != 4 {
		t.Fatalf("CohortSize = %d, want 4", got.CohortSize)
	}
	if len(got.Periods) != 12 {
		t.Fatalf("periods = %d, want 12", len(got.Periods))
	}

	eligible := []int{4, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	renewed := []int{2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	for i, p := range got.Periods {
		if p.Period != i+1 {
			t.Errorf("period %d: Period = %d", i, p.Period)
		}
		if p.Eligible != eligible[i] {
			t.Errorf("period %d: Eligible = %d, want %d", i+1, p.Eligible, eligible[i])
		}
		if p.Renewed != renewed[i] {
			t.Errorf("period %d: Renewed = %d, want %d", i+1, p.Renewed, renewed[i])
		}

		wantRate := float64(renewed[i]) / float64(eligible[i])
		if p.Rate != wantRate {
			t.Errorf("period %d: Rate = %v, want %v", i+1, p.Rate, wantRate)
		}
		if p.CumulativeRate != float64(renewed[i])/4 {
			t.Errorf("period %d: CumulativeRate = %v, want %v", i+1, p.CumulativeRate, float64(renewed[i])/4)
		}
	}
}

func TestRetentionFromUsersUnfinishedPeriod(t *testing.T) {
	users := retentionFixture()
	now := testBase().Add(45 * 24 * time.Hour)

	got := retentionFromUsers(users, 1, now, 12)

	want := cache.RetentionPeriodPayload{
		Period:         1,
		Eligible:       4,
		Renewed:        2,
		Rate:           0.5,
		CumulativeRate: 0.5,
	}
	if len(got.Periods) != 1 || !reflect.DeepEqual(got.Periods[0], want) {
		t.Errorf("periods = %+v, want [%+v]", got.Periods, want)
	}
}

func TestRetentionFromUsersNothingEnded(t *testing.T) {
	users := retentionFixture()
	now := testBase().Add(25 * 24 * time.Hour)

	got := retentionFromUsers(users, 1, now, 12)

	if got.CohortSize != 4 || len(got.Periods) != 0 {
		t.Errorf("retention = %+v, want no periods", got)
	}
}

func TestRetentionFromUsersMaxPeriods(t *testing.T) {
	users := retentionFixture()
	now := testBase().Add(400 * 24 * time.Hour)

	got := retentionFromUsers(users, 1, now, 2)

	if len(got.Periods) != 2 {
		t.Fatalf("periods = %d, want 2", len(got.Periods))
	}
	if got.Periods[1].Eligible != 2 || got.Periods[1].Renewed != 1 {
		t.Errorf("period 2 = %+v, want eligible 2 renewed 1", got.Periods[1])
	}
}

func TestRetentionFromUsersAnnualPlan(t *testing.T) {
	base := testBase()
	users := []cache.UserMetrics{
		{
			UserId:          "annual",
			FirstPlanId:     2,
			FirstPlanMonths: 12,
			FirstSubDate:    base,
			Chain: []cache.ChainItem{
				{PlanId: 2, PeriodStart: base, PeriodEnd: base.Add(12 * testDay)},
			},
		},
	}

	got := retentionFromUsers(users, 2, base.Add(400*24*time.Hour), 12)

	if got.CohortSize != 1 {
		t.Fatalf("CohortSize = %d, want 1", got.CohortSize)
	}
	want := cache.RetentionPeriodPayload{
		Period:   1,
		Eligible: 1,
		Renewed:  0,
		Rate:     0,
	}
	if len(got.Periods) != 1 || !reflect.DeepEqual(got.Periods[0], want) {
		t.Errorf("periods = %+v, want [%+v]", got.Periods, want)
	}
}

func TestLtvByPlanFromUsers(t *testing.T) {
	base := testBase()
	now := base.Add(24 * time.Hour)
	users := []cache.UserMetrics{
		{UserId: "a", FirstPlanId: 1, FirstSubDate: base, Ltv: 1000},
		{UserId: "b", FirstPlanId: 1, FirstSubDate: base, Ltv: 0},
		{UserId: "c", FirstPlanId: 2, FirstSubDate: base, Ltv: 500},
		{UserId: "d", FirstSubDate: base, Ltv: 700},
	}

	got := ltvByPlanFromUsers(users, now)

	want := cache.LtvByPlanPayload{
		DataAsOf: now,
		Plans: []cache.PlanLtvPayload{
			{PlanId: 1, Users: 2, AvgLtv: 500, TotalRevenue: 1000},
			{PlanId: 2, Users: 1, AvgLtv: 500, TotalRevenue: 500},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("by-plan = %+v, want %+v", got, want)
	}
}

type fakeCache struct {
	summary    cache.SummaryPayload
	retentions []cache.RetentionPayload
	snapshot   cache.UserMetricsSnapshot
}

func (f *fakeCache) SaveSummary(_ context.Context, p cache.SummaryPayload) error {
	f.summary = p
	return nil
}

func (f *fakeCache) SaveRetention(_ context.Context, p []cache.RetentionPayload) error {
	f.retentions = p
	return nil
}

func (f *fakeCache) SaveUserMetrics(_ context.Context, p cache.UserMetricsSnapshot) error {
	f.snapshot = p
	return nil
}

func (f *fakeCache) GetUserMetrics(context.Context) (cache.UserMetricsSnapshot, error) {
	return f.snapshot, nil
}

func (f *fakeCache) SaveComputed(context.Context, string, any) error { return nil }

func (f *fakeCache) GetComputed(context.Context, string, any) error { return cache.ErrNotFound }

func TestRecalculateSnapshot(t *testing.T) {
	base := testBase()
	now := base.Add(30 * 24 * time.Hour)

	users := map[string]domain.User{"u1": {}}
	payments := map[string]domain.Payment{
		"pay1": {Id: "pay1", UserId: "u1", Amount: 500, Status: "succeeded"},
		"pay2": {Id: "pay2", UserId: "u1", Amount: 300, Status: "refunded"},
	}
	subs := map[string]domain.Subscription{
		"sub1": {Id: "sub1", UserId: "u1", PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
	}

	fc := &fakeCache{}
	svc := NewMetricsService(fc)

	if err := svc.Recalculate(context.Background(), users, payments, subs, testPlans(), now); err != nil {
		t.Fatalf("Recalculate: %v", err)
	}

	if fc.snapshot.GeneratedAt != now {
		t.Errorf("GeneratedAt = %v, want %v", fc.snapshot.GeneratedAt, now)
	}
	if len(fc.snapshot.Users) != 1 {
		t.Fatalf("users = %d, want 1", len(fc.snapshot.Users))
	}

	u := fc.snapshot.Users[0]
	want := cache.UserMetrics{
		UserId:          "u1",
		Ltv:             500,
		FirstPlanId:     1,
		FirstPlanMonths: 1,
		FirstSubDate:    base,
		Chain: []cache.ChainItem{
			{PlanId: 1, PeriodStart: base, PeriodEnd: base.Add(testDay)},
		},
	}
	if !reflect.DeepEqual(u, want) {
		t.Errorf("user = %+v, want %+v", u, want)
	}
}
