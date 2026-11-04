package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	EnvKey    = "ENV"
	Prod      = "prod"
	prodFile  = ".config/prod.json"
	localFile = ".config/local.json"
)

func LoadConfig[T any](cfg *T) error {
	path := pickPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: read %q: %w", path, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("config: unmarshal %q: %w", path, err)
	}

	return nil
}

func pickPath() string {
	env := strings.ToLower(strings.TrimSpace(os.Getenv(EnvKey)))
	if env == Prod {
		return prodFile
	}
	return localFile
}
