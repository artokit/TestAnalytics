package config

import "shared/config"

type DatabaseConfig struct {
	ConnectionString string
}

type Config struct {
	Address      string `env:"USERS_ADDRESS" envDefault:"0.0.0.0"`
	Port         int    `env:"USERS_PORT" envDefault:"80"`
	DatabaseName string `env:"USERS_DB_NAME" envDefault:"users-service"`
	Database     DatabaseConfig
}

func New() (*Config, error) {
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		return nil, err
	}

	cfg.Database.ConnectionString = config.PostgresDSN(cfg.DatabaseName)

	return &cfg, nil
}
