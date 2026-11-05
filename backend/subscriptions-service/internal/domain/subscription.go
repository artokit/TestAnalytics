package domain

import "time"

// Subscription - строка таблицы subscriptions (миграция 00001).
type Subscription struct {
	Id          string    `db:"id"`
	UserId      string    `db:"user_id"`
	PlanId      int16     `db:"plan_id"`
	PeriodStart time.Time `db:"period_start"`
	PeriodEnd   time.Time `db:"period_end"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}