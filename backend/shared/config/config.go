package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var envOnce sync.Once

func String(key, fallback string) string {
	loadEnvFile()
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func Int(key string, fallback int) (int, error) {
	value, err := parse(key, func(raw string) (int, error) {
		return strconv.Atoi(raw)
	}, fallback)
	return value, err
}

func Int64(key string, fallback int64) (int64, error) {
	value, err := parse(key, func(raw string) (int64, error) {
		return strconv.ParseInt(raw, 10, 64)
	}, fallback)
	return value, err
}

func Bool(key string, fallback bool) (bool, error) {
	value, err := parse(key, strconv.ParseBool, fallback)
	return value, err
}

func Duration(key string, fallback time.Duration) (time.Duration, error) {
	value, err := parse(key, time.ParseDuration, fallback)
	return value, err
}

func parse[T any](key string, fn func(string) (T, error), fallback T) (T, error) {
	raw, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	value, err := fn(raw)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("config: %s: %w", key, err)
	}
	return value, nil
}

func lookup(key string) (string, bool) {
	loadEnvFile()
	value := os.Getenv(key)
	if value == "" {
		return "", false
	}
	return value, true
}

func loadEnvFile() {
	envOnce.Do(func() {
		path := findEnvFile()
		if path == "" {
			return
		}

		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			key, value, ok := parseEnvLine(scanner.Text())
			if !ok || os.Getenv(key) != "" {
				continue
			}
			_ = os.Setenv(key, value)
		}
	})
}

func parseEnvLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")

	key, value, found := strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	if !found || key == "" {
		return "", "", false
	}

	value = strings.TrimSpace(value)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		value = value[1 : len(value)-1]
	}

	return key, value, true
}

func findEnvFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		path := filepath.Join(dir, ".env")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func PostgresDSN(dbName string) string {
	user := String("POSTGRES_USER", "postgres")
	password := String("POSTGRES_PASSWORD", "postgres")
	host := String("POSTGRES_HOST", "postgres")
	port := String("POSTGRES_PORT", "5432")

	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", user, password, host, port, dbName)
}
