package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func ReadIntQuery(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return value, nil
}

func ReadTimeQuery(r *http.Request, key string) (*time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
