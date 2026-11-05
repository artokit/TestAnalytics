package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/pkg"
)

type UserService struct {
	hc     *pkg.HttpClient
	logger *slog.Logger
	config *config.Config
}

func NewUserService(hc *pkg.HttpClient, logger *slog.Logger, cfg *config.Config) *UserService {
	return &UserService{
		hc:     hc,
		logger: logger,
		config: cfg,
	}
}

func (s *UserService) FetchUsers(ctx context.Context, updatedSince *time.Time) ([]domain.User, error) {
	const pageSize = 100

	users := make([]domain.User, 0)

	for page := 1; ; page++ {
		u := fmt.Sprintf("%s/users?page=%d&page_size=%d", s.config.Sync.UsersURL, page, pageSize)
		if updatedSince != nil {
			u += "&updated_since=" + updatedSince.Format(time.RFC3339)
		}

		p, err := pkg.RequestWithRetry(ctx, func() (domain.UsersPage, error) {
			resp, err := s.hc.Get[domain.UsersPage](ctx, u)
			if err != nil {
				return domain.UsersPage{}, err
			}
			if resp.StatusCode != http.StatusOK {
				return domain.UsersPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
			return resp.Body, nil
		})
		if err != nil {
			s.logger.Error("failed to fetch data from users-service", "error", err)
			return nil, fmt.Errorf("failed to fetch data from users-service")
		}

		users = append(users, p.Users...)
		if page >= p.TotalPages {
			return users, nil
		}
	}
}
