package config

import "shared/config"

type Config struct {
	Address  string         `json:"address"`
	Port     int            `json:"port"`
	Database DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
	ConnectionString string `json:"connection_string"`
}

func New() (*Config, error) {
	var cfg Config
	if err := config.LoadConfig[Config](&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
