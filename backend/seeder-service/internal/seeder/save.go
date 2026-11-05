package seeder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"seeder-service/internal/generator"
)

// columns - количество параметров на строку в multi-row INSERT.
const columns = 4 // id, telegram_id, created_at, utm_link

// Save заливает пользователей бачами multi-row INSERT в одной транзакции.
// Строки идут отсортированными по created_at (UUIDv7 монотонен), поэтому
// PK-индекс наполняется последовательно и не расщепляет страницы.
func Save(ctx context.Context, db *sql.DB, users []generator.User, batchSize int) (batches int, err error) {
	if len(users) == 0 {
		return 0, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("seeder: begin tx: %w", err)
	}

	defer func() {
		// ErrTxDone после успешного Commit - нормальная ситуация, не блокируем.
		if pErr := tx.Rollback(); pErr != nil && !errors.Is(pErr, sql.ErrTxDone) && err == nil {
			err = fmt.Errorf("seeder: rollback: %w", pErr)
		}
	}()

	if err = insertUsers(ctx, tx, users, batchSize, &batches); err != nil {
		return batches, err
	}

	if err = tx.Commit(); err != nil {
		return batches, fmt.Errorf("seeder: commit: %w", err)
	}

	return batches, nil
}

func insertUsers(ctx context.Context, tx *sql.Tx, users []generator.User, batchSize int, batches *int) error {
	for start := 0; start < len(users); start += batchSize {
		end := min(start+batchSize, len(users))
		if err := insertUsersBatch(ctx, tx, users[start:end]); err != nil {
			return err
		}
		*batches++
	}
	return nil
}

// insertUsersBatch вставляет батч одной multi-row INSERT.
func insertUsersBatch(ctx context.Context, tx *sql.Tx, batch []generator.User) error {
	args := make([]any, 0, len(batch)*columns)
	for _, user := range batch {
		args = append(args, user.Id, user.TelegramId, user.CreatedAt, user.UtmLink)
	}

	if _, err := tx.ExecContext(ctx, buildUsersInsert(len(batch)), args...); err != nil {
		return fmt.Errorf("seeder: insert batch (%d rows): %w", len(batch), err)
	}

	return nil
}

// buildUsersInsert: INSERT INTO users (...) VALUES ($1,...$4),...
func buildUsersInsert(batchLen int) string {
	return buildMultiRowInsert("users", []string{"id", "telegram_id", "created_at", "utm_link"}, batchLen)
}

// --- subscriptions ---

const subsColumns = 8 // id, user_id, plan_id, period_start, period_end, status, created_at, updated_at

// SaveSubscriptions заливает подписки бачами в одной транзакции.
// overwrite=true -> TRUNCATE перед вставкой (требование FORCE=1).
func SaveSubscriptions(ctx context.Context, db *sql.DB, subs []generator.Subscription, batchSize int, overwrite bool) (batches int, err error) {
	if len(subs) == 0 {
		return 0, nil
	}

	// UUIDv7 монотонен по created_at - сортируем для аккуратной вставки.
	sort.Slice(subs, func(i, j int) bool { return subs[i].CreatedAt.Before(subs[j].CreatedAt) })

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("seeder: begin tx: %w", err)
	}
	defer func() {
		if pErr := tx.Rollback(); pErr != nil && !errors.Is(pErr, sql.ErrTxDone) && err == nil {
			err = fmt.Errorf("seeder: rollback: %w", pErr)
		}
	}()

	if overwrite {
		if _, err = tx.ExecContext(ctx, "TRUNCATE TABLE subscriptions"); err != nil {
			return 0, fmt.Errorf("seeder: truncate subscriptions: %w", err)
		}
	}

	for start := 0; start < len(subs); start += batchSize {
		end := min(start+batchSize, len(subs))
		if err = insertSubsBatch(ctx, tx, subs[start:end]); err != nil {
			return batches, err
		}
		batches++
	}

	if err = tx.Commit(); err != nil {
		return batches, fmt.Errorf("seeder: commit: %w", err)
	}

	return batches, nil
}

func insertSubsBatch(ctx context.Context, tx *sql.Tx, batch []generator.Subscription) error {
	args := make([]any, 0, len(batch)*subsColumns)
	for _, s := range batch {
		args = append(args, s.Id, s.UserId, s.PlanId, s.PeriodStart, s.PeriodEnd, s.Status, s.CreatedAt, s.UpdatedAt)
	}

	if _, err := tx.ExecContext(ctx,
		buildMultiRowInsert("subscriptions", []string{"id", "user_id", "plan_id", "period_start", "period_end", "status", "created_at", "updated_at"}, len(batch)),
		args...); err != nil {
		return fmt.Errorf("seeder: insert subscriptions batch (%d rows): %w", len(batch), err)
	}

	return nil
}

// --- payments ---

const paysColumns = 10 // id, user_id, subscription_id, amount, currency, status, period_start, period_end, created_at, updated_at

// SavePayments заливает платежи бачами в одной транзакции.
func SavePayments(ctx context.Context, db *sql.DB, pays []generator.Payment, batchSize int, overwrite bool) (batches int, err error) {
	if len(pays) == 0 {
		return 0, nil
	}

	sort.Slice(pays, func(i, j int) bool { return pays[i].CreatedAt.Before(pays[j].CreatedAt) })

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("seeder: begin tx: %w", err)
	}
	defer func() {
		if pErr := tx.Rollback(); pErr != nil && !errors.Is(pErr, sql.ErrTxDone) && err == nil {
			err = fmt.Errorf("seeder: rollback: %w", pErr)
		}
	}()

	if overwrite {
		if _, err = tx.ExecContext(ctx, "TRUNCATE TABLE payments"); err != nil {
			return 0, fmt.Errorf("seeder: truncate payments: %w", err)
		}
	}

	for start := 0; start < len(pays); start += batchSize {
		end := min(start+batchSize, len(pays))
		if err = insertPaysBatch(ctx, tx, pays[start:end]); err != nil {
			return batches, err
		}
		batches++
	}

	if err = tx.Commit(); err != nil {
		return batches, fmt.Errorf("seeder: commit: %w", err)
	}

	return batches, nil
}

func insertPaysBatch(ctx context.Context, tx *sql.Tx, batch []generator.Payment) error {
	args := make([]any, 0, len(batch)*paysColumns)
	for _, p := range batch {
		args = append(args, p.Id, p.UserId, p.SubscriptionId, p.Amount, p.Currency, p.Status, p.PeriodStart, p.PeriodEnd, p.CreatedAt, p.UpdatedAt)
	}

	if _, err := tx.ExecContext(ctx,
		buildMultiRowInsert("payments", []string{"id", "user_id", "subscription_id", "amount", "currency", "status", "period_start", "period_end", "created_at", "updated_at"}, len(batch)),
		args...); err != nil {
		return fmt.Errorf("seeder: insert payments batch (%d rows): %w", len(batch), err)
	}

	return nil
}

// --- общее ---

// buildMultiRowInsert: INSERT INTO table (cols...) VALUES ($1..$k),...
// Плейсхолдеры нумеруются в пределах стейтмента: каждый Exec - отдельный батч.
// Лимит протокола учтён в config.maxBatchSize.
func buildMultiRowInsert(table string, cols []string, batchLen int) string {
	n := len(cols)
	var sb strings.Builder

	sb.WriteString("INSERT INTO " + table + " (")
	for i, c := range cols {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(c)
	}
	sb.WriteString(") VALUES ")

	for i := 0; i < batchLen; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for j := 0; j < n; j++ {
			if j > 0 {
				sb.WriteString(",")
			}
			fmt.Fprintf(&sb, "$%d", i*n+j+1)
		}
		sb.WriteString(")")
	}

	return sb.String()
}

// CountTable - сколько строк уже в таблице (для логики idемпотентного запуска).
func CountTable(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var n int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
		return 0, fmt.Errorf("seeder: count %s: %w", table, err)
	}
	return n, nil
}
