package middleware

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvFaultsErrorRate500 = "MOCK_ERROR_RATE_500"
	EnvFaultsErrorRate429 = "MOCK_ERROR_RATE_429"
	EnvFaultsRetryAfter   = "MOCK_RETRY_AFTER"
)

const DefaultFaultsRetryAfter = 5 * time.Second

type FaultsConfig struct {
	ErrorRate500 float64
	ErrorRate429 float64
	RetryAfter   time.Duration
}

func FaultsConfigFromEnv() (FaultsConfig, error) {
	cfg := FaultsConfig{RetryAfter: DefaultFaultsRetryAfter}

	var err error
	if cfg.ErrorRate500, err = readFloatEnv(EnvFaultsErrorRate500); err != nil {
		return cfg, err
	}
	if cfg.ErrorRate429, err = readFloatEnv(EnvFaultsErrorRate429); err != nil {
		return cfg, err
	}
	if cfg.RetryAfter, err = readDurationEnv(EnvFaultsRetryAfter); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func readFloatEnv(name string) (float64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, nil
	}

	if percent := strings.TrimSuffix(raw, "%"); percent != raw {
		value, err := strconv.ParseFloat(percent, 64)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", name, err)
		}

		return value / 100, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	if value < 0 || value > 1 {
		return 0, fmt.Errorf("%s: %.2f out of range [0, 1]", name, value)
	}

	return value, nil
}

func readDurationEnv(name string) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, nil
	}

	if millis, err := strconv.Atoi(raw); err == nil {
		return time.Duration(millis) * time.Millisecond, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}

	return value, nil
}
