package config

import "shared/config"

type DatabaseConfig struct {
	ConnectionString string
}

type Config struct {
	Address      string `env:"PAYMENTS_ADDRESS" envDefault:"0.0.0.0"`
	Port         int    `env:"PAYMENTS_PORT" envDefault:"80"`
	DatabaseName string `env:"PAYMENTS_DB_NAME" envDefault:"payments-service"`
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
