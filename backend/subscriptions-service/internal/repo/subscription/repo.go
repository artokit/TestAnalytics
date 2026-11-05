package subscription

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"subscriptions-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	postgres *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		postgres: db,
	}
}

//go:embed "sql/get_subscriptions.sql"
var getSubscriptionsSql string

//go:embed "sql/get_subscriptions_count.sql"
var getSubscriptionsCountSql string

func (r *Repo) GetSubscriptions(ctx context.Context, limit int, offset int, updatedSince *time.Time) ([]domain.Subscription, int64, error) {
	var subs []domain.Subscription

	if err := r.postgres.SelectContext(ctx, &subs, getSubscriptionsSql, limit, offset, updatedSince); err != nil {
		return nil, 0, fmt.Errorf("repo: select subscriptions: %w", err)
	}

	var total int64
	if err := r.postgres.GetContext(ctx, &total, getSubscriptionsCountSql, updatedSince); err != nil {
		return nil, 0, fmt.Errorf("repo: count subscriptions: %w", err)
	}

	return subs, total, nil
}
