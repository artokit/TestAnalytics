package config

import (
	"fmt"

	"shared/config"
)

const (
	defaultSeed      = 42
	defaultUsers     = 10000
	defaultBatchSize = 500
	defaultHistory   = 24
	usersDBDefault   = "user=postgres password=postgres host=localhost port=5432 dbname=users-service sslmode=disable"
	subsDBDefault    = "user=postgres password=postgres host=localhost port=5432 dbname=subscriptions-service sslmode=disable"
	paysDBDefault    = "user=postgres password=postgres host=localhost port=5432 dbname=payments-service sslmode=disable"
)

const (
	maxBatchSize     = 4096
	minHistoryMonths = 18
	maxHistoryMonths = 24
	minUsers         = 10000
)

type Config struct {
	Seed                int64
	Users               int
	BatchSize           int
	HistoryMonths       int
	DatabaseDSN         string
	SubsDatabaseDSN     string
	PaymentsDatabaseDSN string
	Force               bool
}

func FromEnv() (*Config, error) {
	seed, err := config.Int64("SEED", defaultSeed)
	if err != nil {
		return nil, err
	}

	users, err := config.Int("SEED_USERS", defaultUsers)
	if err != nil {
		return nil, err
	}

	batchSize, err := config.Int("SEED_BATCH_SIZE", defaultBatchSize)
	if err != nil {
		return nil, err
	}

	historyMonths, err := config.Int("SEED_HISTORY_MONTHS", defaultHistory)
	if err != nil {
		return nil, err
	}

	force, err := config.Bool("FORCE", false)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Seed:                seed,
		Users:               users,
		BatchSize:           batchSize,
		HistoryMonths:       historyMonths,
		DatabaseDSN:         config.String("DATABASE_DSN", usersDBDefault),
		SubsDatabaseDSN:     config.String("DATABASE_DSN_SUBSCRIPTIONS", subsDBDefault),
		PaymentsDatabaseDSN: config.String("DATABASE_DSN_PAYMENTS", paysDBDefault),
		Force:               force,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Users < minUsers {
		return fmt.Errorf("config: SEED_USERS должен быть >= %d, получено %d", minUsers, c.Users)
	}
	if c.BatchSize < 1 || c.BatchSize > maxBatchSize {
		return fmt.Errorf("config: SEED_BATCH_SIZE должен быть в [1, %d], получено %d", maxBatchSize, c.BatchSize)
	}
	if c.HistoryMonths < minHistoryMonths || c.HistoryMonths > maxHistoryMonths {
		return fmt.Errorf(
			"config: SEED_HISTORY_MONTHS должен быть в [%d, %d], получено %d",
			minHistoryMonths, maxHistoryMonths, c.HistoryMonths,
		)
	}
	return nil
}
