package cache

import "time"

type UserMetrics struct {
	UserId          string      `json:"user_id"`
	Ltv             int64       `json:"ltv"`
	FirstPlanId     int16       `json:"first_plan_id"`
	FirstPlanMonths int         `json:"first_plan_months"`
	FirstSubDate    time.Time   `json:"first_sub_date"`
	Chain           []ChainItem `json:"chain"`
}

type PlanLtvPayload struct {
	PlanId       int16 `json:"plan_id"`
	Users        int   `json:"users"`
	AvgLtv       int64 `json:"avg_ltv"`
	TotalRevenue int64 `json:"total_revenue"`
}

type LtvByPlanPayload struct {
	DataAsOf time.Time        `json:"data_as_of"`
	Plans    []PlanLtvPayload `json:"plans"`
}

type ChainItem struct {
	PlanId      int16     `json:"plan_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type UserMetricsSnapshot struct {
	GeneratedAt time.Time     `json:"generated_at"`
	Users       []UserMetrics `json:"users"`
}

type RetentionPeriodPayload struct {
	Period         int     `json:"period"`
	Eligible       int     `json:"eligible"`
	Renewed        int     `json:"renewed"`
	Rate           float64 `json:"rate"`
	CumulativeRate float64 `json:"cumulative_rate"`
}

type RetentionPayload struct {
	PlanId     int16                    `json:"plan_id"`
	DataAsOf   time.Time                `json:"data_as_of"`
	CohortSize int                      `json:"cohort_size"`
	Periods    []RetentionPeriodPayload `json:"periods"`
}
