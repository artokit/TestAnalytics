package config

import "test_analytics/shared/config"

type DatabaseConfig struct {
	ConnectionString string `json:"connection_string"`
}

type Config struct {
	Address  string         `json:"address"`
	Port     int            `json:"port"`
	Database DatabaseConfig `json:"database"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := config.LoadConfig[Config](&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
