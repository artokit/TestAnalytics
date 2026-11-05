package payment

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"payments-service/internal/domain"

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

//go:embed "sql/get_payments.sql"
var getPaymentsSql string

//go:embed "sql/get_payments_count.sql"
var getPaymentsCountSql string

func (r *Repo) GetPayments(ctx context.Context, limit int, offset int, updatedSince *time.Time) ([]domain.Payment, int64, error) {
	var payments []domain.Payment
	if err := r.postgres.SelectContext(ctx, &payments, getPaymentsSql, updatedSince, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("repo: select payments: %w", err)
	}

	var total int64
	if err := r.postgres.GetContext(ctx, &total, getPaymentsCountSql, updatedSince); err != nil {
		return nil, 0, fmt.Errorf("repo: count payments: %w", err)
	}

	return payments, total, nil
}
