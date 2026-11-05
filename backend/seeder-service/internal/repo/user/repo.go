package user

import (
	"context"
	_ "embed"
	"fmt"

	"seeder-service/internal/models"
	"shared/database"
)

//go:embed "sql/count_users.sql"
var countUsersSql string

//go:embed "sql/truncate_users.sql"
var truncateUsersSql string

//go:embed "sql/list_users.sql"
var listUsersSql string

//go:embed "sql/insert_users.sql"
var insertUsersSql string

type Repo struct {
	storage *database.Storage
}

func NewRepo(storage *database.Storage) *Repo {
	return &Repo{storage: storage}
}

func (r *Repo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.storage.Db.GetContext(ctx, &count, countUsersSql); err != nil {
		return 0, fmt.Errorf("repo: count users: %w", err)
	}
	return count, nil
}

func (r *Repo) Truncate(ctx context.Context) error {
	if _, err := r.storage.Db.ExecContext(ctx, truncateUsersSql); err != nil {
		return fmt.Errorf("repo: truncate users: %w", err)
	}
	return nil
}

func (r *Repo) List(ctx context.Context) ([]models.User, error) {
	rows, err := r.storage.Db.QueryContext(ctx, listUsersSql)
	if err != nil {
		return nil, fmt.Errorf("repo: select users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.Id, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("repo: scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: iterate users: %w", err)
	}

	return users, nil
}

func (r *Repo) Save(ctx context.Context, users []models.User, batchSize int) (int, error) {
	batches := 0
	for start := 0; start < len(users); start += batchSize {
		end := min(start+batchSize, len(users))
		if err := r.insertBatch(ctx, users[start:end]); err != nil {
			return batches, err
		}
		batches++
	}

	return batches, nil
}

func (r *Repo) insertBatch(ctx context.Context, batch []models.User) error {
	tx, err := r.storage.Db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo: begin tx: %w", err)
	}

	if _, err := tx.NamedExecContext(ctx, insertUsersSql, batch); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("repo: insert users: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo: commit: %w", err)
	}

	return nil
}
