package services

import (
	"context"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/pkg"
)

type SubscriptionService struct {
	hc     *httpclient.HttpClient
	config *config.Config
}

func NewSubscriptionService(hc *httpclient.HttpClient, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		hc:     hc,
		config: cfg,
	}
}

func (s *SubscriptionService) GetSubscriptionsPage(ctx context.Context, url string) (httpclient.Response[domain.SubscriptionsPage], error) {
	return s.hc.Get[domain.SubscriptionsPage](ctx, url)
}
