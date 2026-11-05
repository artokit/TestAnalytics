package payment

import (
	"context"
	_ "embed"
	"fmt"

	"seeder-service/internal/models"
	"shared/database"
)

//go:embed "sql/count_payments.sql"
var countPaymentsSql string

//go:embed "sql/truncate_payments.sql"
var truncatePaymentsSql string

//go:embed "sql/insert_payments.sql"
var insertPaymentsSql string

type Repo struct {
	storage *database.Storage
}

func NewRepo(storage *database.Storage) *Repo {
	return &Repo{storage: storage}
}

func (r *Repo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.storage.Db.GetContext(ctx, &count, countPaymentsSql); err != nil {
		return 0, fmt.Errorf("repo: count payments: %w", err)
	}
	return count, nil
}

func (r *Repo) Truncate(ctx context.Context) error {
	if _, err := r.storage.Db.ExecContext(ctx, truncatePaymentsSql); err != nil {
		return fmt.Errorf("repo: truncate payments: %w", err)
	}
	return nil
}

func (r *Repo) Save(ctx context.Context, pays []models.Payment, batchSize int) (int, error) {
	batches := 0
	for start := 0; start < len(pays); start += batchSize {
		end := min(start+batchSize, len(pays))
		if err := r.insertBatch(ctx, pays[start:end]); err != nil {
			return batches, err
		}
		batches++
	}

	return batches, nil
}

func (r *Repo) insertBatch(ctx context.Context, batch []models.Payment) error {
	tx, err := r.storage.Db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo: begin tx: %w", err)
	}

	if _, err := tx.NamedExecContext(ctx, insertPaymentsSql, batch); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("repo: insert payments: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo: commit: %w", err)
	}

	return nil
}
