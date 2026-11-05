package services

import (
	"context"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/pkg"
)

type PaymentService struct {
	hc     *httpclient.HttpClient
	config *config.Config
}

func NewPaymentService(hc *httpclient.HttpClient, cfg *config.Config) *PaymentService {
	return &PaymentService{
		hc:     hc,
		config: cfg,
	}
}

func (s *PaymentService) GetPaymentsPage(ctx context.Context, url string) (httpclient.Response[domain.PaymentsPage], error) {
	return s.hc.Get[domain.PaymentsPage](ctx, url)
}
