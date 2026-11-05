package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	envFileName = ".env"
	envTag      = "env"
	skipTag     = "-"
	defaultTag  = "envDefault"
)

type duration = time.Duration

func Load(target any) error {
	loadEnvFile()

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: target must be a pointer to struct")
	}

	return fill(value.Elem())
}

func fill(value reflect.Value) error {
	t := value.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		target := value.Field(i)
		key := strings.TrimSpace(field.Tag.Get(envTag))

		if key == skipTag {
			continue
		}
		if key != "" {
			raw := os.Getenv(key)
			if raw == "" {
				raw = field.Tag.Get(defaultTag)
			}
			if err := parse(target, raw); err != nil {
				return fmt.Errorf("config: %s: %w", key, err)
			}
			continue
		}
		if target.Kind() != reflect.Struct {
			continue
		}
		if err := fill(target); err != nil {
			return err
		}
	}

	return nil
}

func parse(field reflect.Value, raw string) error {
	if field.Type() == reflect.TypeOf(duration(0)) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		field.SetInt(int64(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(v)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}

	return nil
}

func loadEnvFile() {
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
		path := filepath.Join(dir, envFileName)
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

func Getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func PostgresDSN(dbName string) string {
	user := Getenv("POSTGRES_USER", "postgres")
	password := Getenv("POSTGRES_PASSWORD", "postgres")
	host := Getenv("POSTGRES_HOST", "postgres")
	port := Getenv("POSTGRES_PORT", "5432")

	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", user, password, host, port, dbName)
}
