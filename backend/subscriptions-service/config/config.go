package config

import "shared/config"

type DatabaseConfig struct {
	ConnectionString string
}

type Config struct {
	Address      string `env:"SUBSCRIPTIONS_ADDRESS" envDefault:"0.0.0.0"`
	Port         int    `env:"SUBSCRIPTIONS_PORT" envDefault:"80"`
	DatabaseName string `env:"SUBSCRIPTIONS_DB_NAME" envDefault:"subscriptions-service"`
	Database     DatabaseConfig
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		return nil, err
	}

	cfg.Database.ConnectionString = config.PostgresDSN(cfg.DatabaseName)

	return &cfg, nil
}
