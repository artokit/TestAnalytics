package request

import (
	"net/http"
	"payments-service/internal/errors"
	"time"

	"test_analytics/shared/httputil"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 50
	MinPage         = 1
	MinPageSize     = 1
)

type PaymentFilter struct {
	Page         int
	PageSize     int
	UpdatedSince *time.Time
}

func GetPaymentFilter(r *http.Request) (*PaymentFilter, error) {
	page, err := httputil.ReadIntQuery(r, "page", DefaultPage)
	if err != nil {
		return nil, errors.ErrInvalidQueryParameter
	}

	pageSize, err := httputil.ReadIntQuery(r, "page_size", DefaultPageSize)
	if err != nil {
		return nil, errors.ErrInvalidQueryParameter
	}

	updatedSince, err := httputil.ReadTimeQuery(r, "updated_since")
	if err != nil {
		return nil, errors.ErrInvalidQueryParameter
	}

	return &PaymentFilter{
		Page:         page,
		PageSize:     pageSize,
		UpdatedSince: updatedSince,
	}, nil
}
