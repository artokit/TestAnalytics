package config

import (
	"fmt"
	"time"

	"shared/config"
)

type RedisConfig struct {
	Address string `env:"REDIS_ADDRESS" envDefault:"redis:6379"`
}

type SyncConfig struct {
	UsersURL         string `env:"SYNC_USERS_URL" envDefault:"http://users-service:80"`
	PaymentsURL      string `env:"SYNC_PAYMENTS_URL" envDefault:"http://payments-service:80"`
	SubscriptionsURL string `env:"SYNC_SUBSCRIPTIONS_URL" envDefault:"http://subscriptions-service:80"`
	IntervalRaw      string `env:"SYNC_INTERVAL"`
	Interval         time.Duration
}

type Config struct {
	Address string      `env:"ANALYTICS_ADDRESS" envDefault:"0.0.0.0"`
	Port    int         `env:"ANALYTICS_PORT" envDefault:"80"`
	Redis   RedisConfig
	Sync    SyncConfig
}

const defaultSyncInterval = 10 * time.Minute

func NewConfig() (*Config, error) {
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		return nil, err
	}

	cfg.Sync.Interval = defaultSyncInterval
	if cfg.Sync.IntervalRaw != "" {
		interval, err := time.ParseDuration(cfg.Sync.IntervalRaw)
		if err != nil {
			return nil, fmt.Errorf("config: SYNC_INTERVAL: %w", err)
		}
		cfg.Sync.Interval = interval
	}

	return &cfg, nil
}
