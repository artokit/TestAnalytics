package config

import (
	"fmt"
	"os"
	"time"

	"shared/config"
)

type RedisConfig struct {
	Address string `json:"address"`
}

type Config struct {
	Address string      `json:"address"`
	Port    int         `json:"port"`
	Redis   RedisConfig `json:"redis"`
	Sync    SyncConfig  `json:"sync"`
}

// SyncConfig - настройки фонового воркера синхронизации.
// Interval задаётся через SYNC_INTERVAL (по умолчанию 10m).
type SyncConfig struct {
	UsersURL         string        `json:"users_url"`          // адрес users-service (GET /users)
	PaymentsURL      string        `json:"payments_url"`       // адрес payments-service (GET /payments)
	SubscriptionsURL string        `json:"subscriptions_url"`  // адрес subscriptions-service (GET /subscriptions)
	PlansURL         string        `json:"plans_url"`          // адрес subscriptions-service (GET /plans)
	IntervalRaw      string        `json:"interval"`           // сырой интервал из конфига
	Interval         time.Duration `json:"-"`                  // распарсенный интервал
}

const defaultSyncInterval = 10 * time.Minute

func NewConfig() (*Config, error) {
	var cfg Config
	if err := config.LoadConfig[Config](&cfg); err != nil {
		return nil, err
	}


	intervalRaw := cfg.Sync.IntervalRaw
	if v := os.Getenv("SYNC_INTERVAL"); v != "" {
		intervalRaw = v
	}
	if intervalRaw == "" {
		cfg.Sync.Interval = defaultSyncInterval
	} else {
		interval, err := time.ParseDuration(intervalRaw)
		if err != nil {
			return nil, fmt.Errorf("config: SYNC_INTERVAL: %w", err)
		}
		cfg.Sync.Interval = interval
	}

	return &cfg, nil
}
