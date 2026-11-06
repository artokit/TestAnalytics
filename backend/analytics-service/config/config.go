package config

import (
	"time"

	"shared/config"
)

type RedisConfig struct {
	Address string
}

type Config struct {
	Address string
	Port    int
	Redis   RedisConfig
	Sync    SyncConfig
}

type SyncConfig struct {
	UsersURL         string
	PaymentsURL      string
	SubscriptionsURL string
	Interval         time.Duration
}

const defaultSyncInterval = 10 * time.Minute

func NewConfig() (*Config, error) {
	port, err := config.Int("ANALYTICS_PORT", 80)
	if err != nil {
		return nil, err
	}

	interval, err := config.Duration("SYNC_INTERVAL", defaultSyncInterval)
	if err != nil {
		return nil, err
	}

	return &Config{
		Address: config.String("ANALYTICS_ADDRESS", "0.0.0.0"),
		Port:    port,
		Redis: RedisConfig{
			Address: config.String("REDIS_ADDRESS", "redis:6379"),
		},
		Sync: SyncConfig{
			UsersURL:         config.String("SYNC_USERS_URL", "http://users-service:80"),
			PaymentsURL:      config.String("SYNC_PAYMENTS_URL", "http://payments-service:80"),
			SubscriptionsURL: config.String("SYNC_SUBSCRIPTIONS_URL", "http://subscriptions-service:80"),
			Interval:         interval,
		},
	}, nil
}
