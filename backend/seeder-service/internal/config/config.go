package config

import (
	"fmt"
	"os"
	"strconv"
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
	cfg := &Config{
		Seed:                defaultSeed,
		Users:               defaultUsers,
		BatchSize:           defaultBatchSize,
		HistoryMonths:       defaultHistory,
		DatabaseDSN:         usersDBDefault,
		SubsDatabaseDSN:     subsDBDefault,
		PaymentsDatabaseDSN: paysDBDefault,
	}

	if v := os.Getenv("SEED"); v != "" {
		seed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("config: SEED: %w", err)
		}
		cfg.Seed = seed
	}

	if v := os.Getenv("SEED_USERS"); v != "" {
		users, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: SEED_USERS: %w", err)
		}
		cfg.Users = users
	}

	if v := os.Getenv("SEED_BATCH_SIZE"); v != "" {
		batch, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: SEED_BATCH_SIZE: %w", err)
		}
		cfg.BatchSize = batch
	}

	if v := os.Getenv("SEED_HISTORY_MONTHS"); v != "" {
		months, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: SEED_HISTORY_MONTHS: %w", err)
		}
		cfg.HistoryMonths = months
	}

	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DatabaseDSN = v
	}

	if v := os.Getenv("DATABASE_DSN_SUBSCRIPTIONS"); v != "" {
		cfg.SubsDatabaseDSN = v
	}

	if v := os.Getenv("DATABASE_DSN_PAYMENTS"); v != "" {
		cfg.PaymentsDatabaseDSN = v
	}

	switch v := os.Getenv("FORCE"); v {
	case "":
		cfg.Force = false
	case "1", "true", "True", "TRUE", "yes":
		cfg.Force = true
	case "0", "false", "False", "FALSE", "no":
		cfg.Force = false
	default:
		return nil, fmt.Errorf("config: FORCE: неизвестное значение %q", v)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

const maxBatchSize = 4096

const (
	minHistoryMonths = 18
	maxHistoryMonths = 24
)

const minUsers = 10000

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
