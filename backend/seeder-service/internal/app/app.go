package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"

	"seeder-service/internal/config"
	"seeder-service/internal/generator"
	"seeder-service/internal/models"
	"seeder-service/internal/repo/payment"
	"seeder-service/internal/repo/subscription"
	"seeder-service/internal/repo/user"
	"shared/database"
)

func Run(ctx context.Context) error {
	started := time.Now()

	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "seeder")

	logger.Info("seeder: старт",
		"seed", cfg.Seed,
		"users", cfg.Users,
		"batch_size", cfg.BatchSize,
		"history_months", cfg.HistoryMonths,
		"force", cfg.Force,
	)

	usersDB, err := database.New(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open users db: %w", err)
	}
	defer func() { _ = usersDB.Close() }()

	if err := usersDB.Db.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping users db: %w", err)
	}

	subsDB, err := database.New(cfg.SubsDatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open subscriptions db: %w", err)
	}
	defer func() { _ = subsDB.Close() }()

	if err := subsDB.Db.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping subscriptions db: %w", err)
	}

	paysDB, err := database.New(cfg.PaymentsDatabaseDSN)
	if err != nil {
		return fmt.Errorf("seeder: open payments db: %w", err)
	}
	defer func() { _ = paysDB.Close() }()

	if err := paysDB.Db.PingContext(ctx); err != nil {
		return fmt.Errorf("seeder: ping payments db: %w", err)
	}

	userRepo := user.NewRepo(usersDB)
	subsRepo := subscription.NewRepo(subsDB)
	paysRepo := payment.NewRepo(paysDB)

	if err := seedUsers(ctx, cfg, logger, userRepo); err != nil {
		return err
	}

	users, err := userRepo.List(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("seeder: таблица users пуста, подписки не на чем строить")
	}

	if err := seedHistory(ctx, cfg, logger, subsRepo, paysRepo, users); err != nil {
		return err
	}

	logger.Info("seeder: готово", "duration", time.Since(started).String())
	return nil
}

func seedUsers(ctx context.Context, cfg *config.Config, logger *slog.Logger, repo *user.Repo) error {
	existing, err := repo.Count(ctx)
	if err != nil {
		return err
	}

	if existing > 0 && !cfg.Force {
		logger.Info("seeder: users уже есть, скип (перезаливка - FORCE=1)", "existing_rows", existing)
		return nil
	}

	if existing > 0 && cfg.Force {
		logger.Info("seeder: users не пустая, FORCE=1 -> перезаливаем", "existing_rows", existing)
		if err := repo.Truncate(ctx); err != nil {
			return err
		}
	}

	users, err := generator.GenerateUsers(cfg.Users, cfg.Seed, time.Now(), cfg.HistoryMonths)
	if err != nil {
		return fmt.Errorf("seeder: generate users: %w", err)
	}

	batches, err := repo.Save(ctx, users, cfg.BatchSize)
	if err != nil {
		return err
	}

	logger.Info("seeder: users готовы", "rows", len(users), "batches", batches)
	return nil
}

func seedHistory(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	subsRepo *subscription.Repo,
	paysRepo *payment.Repo,
	users []models.User,
) error {
	subsExisting, err := subsRepo.Count(ctx)
	if err != nil {
		return err
	}
	paysExisting, err := paysRepo.Count(ctx)
	if err != nil {
		return err
	}

	switch {
	case subsExisting == 0 && paysExisting == 0:
	case cfg.Force:
		logger.Info("seeder: таблицы не пустые, FORCE=1 -> перезаливаем",
			"subscriptions", subsExisting, "payments", paysExisting)
	default:
		logger.Info("seeder: история уже есть, скип (перезаливка - FORCE=1)",
			"subscriptions", subsExisting, "payments", paysExisting)
		return nil
	}

	history, err := generator.GenerateHistory(users, cfg.Seed, time.Now())
	if err != nil {
		return fmt.Errorf("seeder: generate history: %w", err)
	}

	if cfg.Force {
		if err := subsRepo.Truncate(ctx); err != nil {
			return err
		}
		if err := paysRepo.Truncate(ctx); err != nil {
			return err
		}
	}

	subsBatches, err := subsRepo.Save(ctx, history.Subs, cfg.BatchSize)
	if err != nil {
		return err
	}

	paysBatches, err := paysRepo.Save(ctx, history.Payments, cfg.BatchSize)
	if err != nil {
		return err
	}

	logger.Info("seeder: история готова",
		"subscriptions", len(history.Subs), "subs_batches", subsBatches,
		"payments", len(history.Payments), "pays_batches", paysBatches,
	)

	return nil
}
