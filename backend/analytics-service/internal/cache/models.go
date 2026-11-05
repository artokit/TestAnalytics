package cache

import "time"

type UserMetrics struct {
	UserId       string      `json:"user_id"`
	Ltv          int64       `json:"ltv"`
	FirstPlanId  int16       `json:"first_plan_id"`
	FirstSubDate time.Time   `json:"first_sub_date"`
	Chain        []ChainItem `json:"chain"`
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
