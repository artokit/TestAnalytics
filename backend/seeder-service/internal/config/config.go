package config

import (
	"fmt"

	"shared/config"
)

const (
	maxBatchSize     = 4096
	minHistoryMonths = 18
	maxHistoryMonths = 24
	minUsers         = 10000
)

type Config struct {
	Seed                int64  `env:"SEED" envDefault:"42"`
	Users               int    `env:"SEED_USERS" envDefault:"10000"`
	BatchSize           int    `env:"SEED_BATCH_SIZE" envDefault:"500"`
	HistoryMonths       int    `env:"SEED_HISTORY_MONTHS" envDefault:"24"`
	DatabaseDSN         string `env:"DATABASE_DSN" envDefault:"user=postgres password=postgres host=localhost port=5432 dbname=users-service sslmode=disable"`
	SubsDatabaseDSN     string `env:"DATABASE_DSN_SUBSCRIPTIONS" envDefault:"user=postgres password=postgres host=localhost port=5432 dbname=subscriptions-service sslmode=disable"`
	PaymentsDatabaseDSN string `env:"DATABASE_DSN_PAYMENTS" envDefault:"user=postgres password=postgres host=localhost port=5432 dbname=payments-service sslmode=disable"`
	Force               bool   `env:"FORCE" envDefault:"false"`
}

func FromEnv() (*Config, error) {
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
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
