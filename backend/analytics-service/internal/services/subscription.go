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

type SubscriptionService struct {
	hc     *pkg.HttpClient
	logger *slog.Logger
	config *config.Config
}

func NewSubscriptionService(hc *pkg.HttpClient, logger *slog.Logger, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		hc:     hc,
		logger: logger,
		config: cfg,
	}
}

func (s *SubscriptionService) FetchSubscriptions(ctx context.Context, updatedSince *time.Time) ([]domain.Subscription, error) {
	const pageSize = 100

	subs := make([]domain.Subscription, 0)

	for page := 1; ; page++ {
		u := fmt.Sprintf("%s/subscriptions?page=%d&page_size=%d", s.config.Sync.SubscriptionsURL, page, pageSize)
		if updatedSince != nil {
			u += "&updated_since=" + updatedSince.Format(time.RFC3339)
		}

		p, err := pkg.RequestWithRetry(ctx, func() (domain.SubscriptionsPage, error) {
			resp, err := s.hc.Get[domain.SubscriptionsPage](ctx, u)
			if err != nil {
				return domain.SubscriptionsPage{}, err
			}
			if resp.StatusCode != http.StatusOK {
				return domain.SubscriptionsPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
			return resp.Body, nil
		})
		if err != nil {
			s.logger.Error("failed to fetch data from subscriptions-service", "error", err)
			return nil, fmt.Errorf("ошибка при получение данных из subscriptions-service")
		}

		subs = append(subs, p.Subscriptions...)
		if page >= p.TotalPages {
			return subs, nil
		}
	}
}

func (s *SubscriptionService) FetchPlans(ctx context.Context) ([]domain.Plan, error) {
	p, err := pkg.RequestWithRetry(ctx, func() (domain.PlansResponse, error) {
		resp, err := s.hc.Get[domain.PlansResponse](ctx, fmt.Sprintf("%s/plans", s.config.Sync.SubscriptionsURL))
		if err != nil {
			return domain.PlansResponse{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return domain.PlansResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return resp.Body, nil
	})
	if err != nil {
		s.logger.Error("failed to fetch plans from subscriptions-service", "error", err)
		return nil, fmt.Errorf("ошибка при получение тарифов из subscriptions-service")
	}

	return p.Items, nil
}
