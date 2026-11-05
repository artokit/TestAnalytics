package response

import (
	"time"
)

type SubscriptionResponse struct {
	Id          string    `json:"id"`
	UserId      string    `json:"user_id"`
	PlanId      int16     `json:"plan_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SubscriptionsResponsePaged struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
	Total         int                    `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	TotalPages    int                    `json:"total_pages"`
}
