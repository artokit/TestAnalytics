package request

import (
	"net/http"
	"time"

	"shared/httputil"
	subscription_errors "subscriptions-service/internal/errors/subscription"
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
		return nil, subscription_errors.ErrInvalidPage
	}

	pageSize, err := httputil.ReadIntQuery(r, "page_size", DefaultPageSize)
	if err != nil {
		return nil, subscription_errors.ErrInvalidPageSize
	}

	updatedSince, err := httputil.ReadTimeQuery(r, "updated_since")
	if err != nil {
		return nil, subscription_errors.ErrInvalidUpdatedSince
	}

	return &SubscriptionFilter{
		Page:         page,
		PageSize:     pageSize,
		UpdatedSince: updatedSince,
	}, nil
}
