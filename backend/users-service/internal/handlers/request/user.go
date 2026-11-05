package request

import (
	"net/http"
	"time"

	"shared/httputil"
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
	page, err := httputil.ReadIntQuery(r, "page", DefaultPage)
	if err != nil {
		return nil, domain.ErrInvalidParam
	}

	pageSize, err := httputil.ReadIntQuery(r, "page_size", DefaultPageSize)
	if err != nil {
		return nil, domain.ErrInvalidParam
	}

	updatedSince, err := httputil.ReadTimeQuery(r, "updated_since")
	if err != nil {
		return nil, domain.ErrInvalidParam
	}

	return &UserFilter{
		Page:         page,
		PageSize:     pageSize,
		UpdatedSince: updatedSince,
	}, nil
}
