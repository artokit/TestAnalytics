package domain

import "time"

type Payment struct {
	Id             string    `db:"id"`
	UserId         string    `db:"user_id"`
	SubscriptionId string    `db:"subscription_id"`
	Amount         int64     `db:"amount"`
	Currency       string    `db:"currency"`
	Status         string    `db:"status"`
	PeriodStart    time.Time `db:"period_start"`
	PeriodEnd      time.Time `db:"period_end"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
