package user

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"users-service/internal/domain"

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

//go:embed "sql/get_users.sql"
var getUsersSql string

//go:embed "sql/get_users_count.sql"
var getUsersCountSql string

func (r *Repo) GetUsers(ctx context.Context, limit int, offset int, updatedSince *time.Time) ([]domain.User, int64, error) {
	var users []domain.User
	if err := r.postgres.SelectContext(ctx, &users, getUsersSql, limit, offset, updatedSince); err != nil {
		return nil, 0, fmt.Errorf("repo: select users: %w", err)
	}

	var total int64
	if err := r.postgres.GetContext(ctx, &total, getUsersCountSql, updatedSince); err != nil {
		return nil, 0, fmt.Errorf("repo: count users: %w", err)
	}

	return users, total, nil
}
