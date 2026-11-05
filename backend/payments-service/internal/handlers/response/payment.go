package response

import (
	"time"
)

type PaymentResponse struct {
	Id             string    `json:"id"`
	UserId         string    `json:"user_id"`
	SubscriptionId string    `json:"subscription_id"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaymentsResponsePaged struct {
	Payments   []PaymentResponse `json:"payments"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}
