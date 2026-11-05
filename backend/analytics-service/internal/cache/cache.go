package cache

import (
	"context"
	"time"
)

type Cache interface {
	SaveSummary(ctx context.Context, summary SummaryPayload) error
	SaveRetention(ctx context.Context, retentions []RetentionPayload) error
	GetSummary(ctx context.Context) (SummaryPayload, bool, error)
	GetRetention(ctx context.Context) ([]RetentionPayload, bool, error)
	SaveUserMetrics(ctx context.Context, snapshot UserMetricsSnapshot) error
	GetUserMetrics(ctx context.Context) (UserMetricsSnapshot, bool, error)
}

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
	GeneratedAt time.Time      `json:"generated_at"`
	Users       []UserMetrics  `json:"users"`
}
