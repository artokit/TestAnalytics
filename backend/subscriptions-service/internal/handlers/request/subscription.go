package request

import (
	"errors"
	"net/http"
	"time"

	"test_analytics/shared/httputil"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 50
	MinPage         = 1
	MinPageSize     = 1
)

type SubscriptionFilter struct {
	Page         int
	PageSize     int
	UpdatedSince *time.Time
}

func GetSubscriptionFilter(r *http.Request) (*SubscriptionFilter, error) {
	page, err := httputil.ReadIntQuery(r, "page", DefaultPage)
	if err != nil {
		return nil, errors.New("неверно передан параметр page")
	}

	pageSize, err := httputil.ReadIntQuery(r, "page_size", DefaultPageSize)
	if err != nil {
		return nil, errors.New("неверно передан параметр page_size")
	}

	updatedSince, err := httputil.ReadTimeQuery(r, "updated_since")
	if err != nil {
		return nil, errors.New("неверно передан параметр updated_since, expected RFC3339")
	}

	return &SubscriptionFilter{
		Page:         page,
		PageSize:     pageSize,
		UpdatedSince: updatedSince,
	}, nil
}
