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

type SyncConfig struct {
	UsersURL         string        `json:"users_url"`
	PaymentsURL      string        `json:"payments_url"`
	SubscriptionsURL string        `json:"subscriptions_url"`
	IntervalRaw      string        `json:"interval"`
	Interval         time.Duration `json:"-"`
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
