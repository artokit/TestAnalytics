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

func NewConfig() (*Config, error) {
	port, err := config.Int("PAYMENTS_PORT", 80)
	if err != nil {
		return nil, err
	}

	return &Config{
		Address: config.String("PAYMENTS_ADDRESS", "0.0.0.0"),
		Port:    port,
		Database: DatabaseConfig{
			ConnectionString: config.PostgresDSN(config.String("PAYMENTS_DB_NAME", "payments-service")),
		},
	}, nil
}
