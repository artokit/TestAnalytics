package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"

	"seeder-service/internal/config"
	"seeder-service/internal/generator"
	"seeder-service/internal/seeder"
)

// usersTable / subscriptionsTable / paymentsTable - имена таблиц для счётчиков.
const (
	usersTable = "users"
	subsTable  = "subscriptions"
	paysTable  = "payments"
)

func Run(ctx context.Context) error {
	started := time.Now()

	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("seeder: старт",
		"seed", cfg.Seed,
		"users", cfg.Users,
		"batch_size", cfg.BatchSize,
		"history_months", cfg.HistoryMonths,
		"force", cfg.Force,
	)

	// --- БД users ---
	usersDB, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open users db: %w", err)
	}
	defer func() { _ = usersDB.Close() }()

	if err = usersDB.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping users db: %w", err)
	}

	if err = seedUsers(ctx, usersDB, cfg, logger); err != nil {
		return err
	}

	// История подписок/платежей строится по юзерам, реально лежащим в users
	// (id + created_at), чтобы ссылки user_id всегда существовали.
	users, err := loadUsers(ctx, usersDB)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("seeder: таблица users пуста, подписки не на чем строить")
	}

	// --- БД subscriptions ---
	subsDB, err := sql.Open("postgres", cfg.SubsDatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open subscriptions db: %w", err)
	}
	defer func() { _ = subsDB.Close() }()

	if err = subsDB.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping subscriptions db: %w", err)
	}

	// --- БД payments ---
	paysDB, err := sql.Open("postgres", cfg.PaymentsDatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open payments db: %w", err)
	}
	defer func() { _ = paysDB.Close() }()

	if err = paysDB.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping payments db: %w", err)
	}

	// Подписки и платежи - один логический датасет: генерируются и заливаются вместе.
	subsExisting, err := seeder.CountTable(ctx, subsDB, subsTable)
	if err != nil {
		return err
	}
	paysExisting, err := seeder.CountTable(ctx, paysDB, paysTable)
	if err != nil {
		return err
	}

	switch {
	case subsExisting == 0 && paysExisting == 0:
		// пустые таблицы - заливаем
	case cfg.Force:
		logger.Info("seeder: таблицы не пустые, FORCE=1 -> перезаливаем",
			"subscriptions", subsExisting, "payments", paysExisting)
	default:
		logger.Info("seeder: история уже есть, скип (перезаливка - FORCE=1)",
			"subscriptions", subsExisting, "payments", paysExisting)
		return nil
	}

	history, stats, err := generator.GenerateHistory(users, cfg.Seed, time.Now())
	if err != nil {
		return fmt.Errorf("seeder: generate history: %w", err)
	}

	subsBatches, err := seeder.SaveSubscriptions(ctx, subsDB, history.Subs, cfg.BatchSize, cfg.Force)
	if err != nil {
		return err
	}

	paysBatches, err := seeder.SavePayments(ctx, paysDB, history.Payments, cfg.BatchSize, cfg.Force)
	if err != nil {
		return err
	}

	logger.Info("seeder: история готова",
		"subscriptions", len(history.Subs), "subs_batches", subsBatches,
		"payments", len(history.Payments), "pays_batches", paysBatches,
		// сводка по пограничным кейсам
		"users_without_payment", stats.UsersWithoutPayment,
		"failed_payments", stats.FailedPayments,
		"retried_payments", stats.RetriedPayments,
		"refunds", stats.Refunds,
		"plan_switches", stats.PlanSwitches,
		"active_now", stats.ActiveNow,
		"renewed", stats.Renewed,
		"expired", stats.Expired,
		"duration", time.Since(started).String(),
	)

	return nil
}

// seedUsers гарантирует наличие пользователей в users-БД:
// пустая -> генерирует и заливает; FORCE=1 -> перезаливает; иначе скип.
func seedUsers(ctx context.Context, db *sql.DB, cfg *config.Config, logger *slog.Logger) error {
	existing, err := seeder.CountTable(ctx, db, usersTable)
	if err != nil {
		return err
	}

	switch {
	case existing == 0:
		// пустая таблица - просто заливаем
	case cfg.Force:
		logger.Info("seeder: users не пустая, FORCE=1 -> перезаливаем", "existing_rows", existing)
		if _, err = db.ExecContext(ctx, "TRUNCATE TABLE "+usersTable); err != nil {
			return fmt.Errorf("seeder: truncate users: %w", err)
		}
	default:
		// идемпотентный запуск: данные уже есть, ничего не делаем
		logger.Info("seeder: users уже есть, скип (перезаливка - FORCE=1)", "existing_rows", existing)
		return nil
	}

	users, err := generator.GenerateUsers(cfg.Users, cfg.Seed, time.Now(), cfg.HistoryMonths)
	if err != nil {
		return fmt.Errorf("seeder: generate users: %w", err)
	}

	batches, err := seeder.Save(ctx, db, users, cfg.BatchSize)
	if err != nil {
		return err
	}

	logger.Info("seeder: users готовы", "rows", len(users), "batches", batches)

	return nil
}

// loadUsers читает id + created_at всех пользователей (нужны для истории).
func loadUsers(ctx context.Context, db *sql.DB) ([]generator.User, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, created_at FROM users ORDER BY created_at ASC")
	if err != nil {
		return nil, fmt.Errorf("seeder: select users: %w", err)
	}
	defer rows.Close()

	users := make([]generator.User, 0)
	for rows.Next() {
		var u generator.User
		if err = rows.Scan(&u.Id, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("seeder: scan user: %w", err)
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("seeder: iterate users: %w", err)
	}

	return users, nil
}
