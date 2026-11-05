package subscription

import (
	"context"
	_ "embed"
	"fmt"

	"seeder-service/internal/models"
	"shared/database"
)

//go:embed "sql/count_subscriptions.sql"
var countSubscriptionsSql string

//go:embed "sql/truncate_subscriptions.sql"
var truncateSubscriptionsSql string

//go:embed "sql/insert_subscriptions.sql"
var insertSubscriptionsSql string

type Repo struct {
	storage *database.Storage
}

func NewRepo(storage *database.Storage) *Repo {
	return &Repo{storage: storage}
}

func (r *Repo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.storage.Db.GetContext(ctx, &count, countSubscriptionsSql); err != nil {
		return 0, fmt.Errorf("repo: count subscriptions: %w", err)
	}
	return count, nil
}

func (r *Repo) Truncate(ctx context.Context) error {
	if _, err := r.storage.Db.ExecContext(ctx, truncateSubscriptionsSql); err != nil {
		return fmt.Errorf("repo: truncate subscriptions: %w", err)
	}
	return nil
}

func (r *Repo) Save(ctx context.Context, subs []models.Subscription, batchSize int) (int, error) {
	batches := 0
	for start := 0; start < len(subs); start += batchSize {
		end := min(start+batchSize, len(subs))
		if err := r.insertBatch(ctx, subs[start:end]); err != nil {
			return batches, err
		}
		batches++
	}

	return batches, nil
}

func (r *Repo) insertBatch(ctx context.Context, batch []models.Subscription) error {
	tx, err := r.storage.Db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo: begin tx: %w", err)
	}

	if _, err := tx.NamedExecContext(ctx, insertSubscriptionsSql, batch); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("repo: insert subscriptions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo: commit: %w", err)
	}

	return nil
}
