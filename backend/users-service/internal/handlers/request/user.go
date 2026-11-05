package request

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"users-service/internal/domain"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 50
	MinPage         = 1
	MinPageSize     = 1
)

type UserFilter struct {
	Page         int
	PageSize     int
	UpdatedSince *time.Time
}

func GetUserFilter(r *http.Request) (*UserFilter, error) {
	page, err := getQueryParamInt(r, "page", DefaultPage)
	if err != nil {
		return nil, domain.ErrInvalidParam
	}

	pageSize, err := getQueryParamInt(r, "page_size", DefaultPageSize)
	if err != nil {
		return nil, domain.ErrInvalidParam
	}

	var updatedSince *time.Time
	if raw := r.URL.Query().Get("updated_since"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, domain.ErrInvalidParam
		}
		updatedSince = &t
	}

	return &UserFilter{
		Page:         page,
		PageSize:     pageSize,
		UpdatedSince: updatedSince,
	}, nil
}

func getQueryParamInt(r *http.Request, key string, fallback int) (int, error) {
	query := r.URL.Query()
	if query.Get(key) == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(query.Get(key))
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return value, nil
}
