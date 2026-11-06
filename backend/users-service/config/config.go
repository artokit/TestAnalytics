package config

import "shared/config"

type DatabaseConfig struct {
	ConnectionString string
}

type Config struct {
	Address  string
	Port     int
	Database DatabaseConfig
}

func New() (*Config, error) {
	port, err := config.Int("USERS_PORT", 80)
	if err != nil {
		return nil, err
	}

	return &Config{
		Address: config.String("USERS_ADDRESS", "0.0.0.0"),
		Port:    port,
		Database: DatabaseConfig{
			ConnectionString: config.PostgresDSN(config.String("USERS_DB_NAME", "users-service")),
		},
	}, nil
}
