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

type PaymentService struct {
	hc     *pkg.HttpClient
	logger *slog.Logger
	config *config.Config
}

func NewPaymentService(hc *pkg.HttpClient, logger *slog.Logger, cfg *config.Config) *PaymentService {
	return &PaymentService{
		hc:     hc,
		logger: logger,
		config: cfg,
	}
}

func (s *PaymentService) FetchPayments(ctx context.Context, updatedSince *time.Time) ([]domain.Payment, error) {
	const pageSize = 100

	payments := make([]domain.Payment, 0)

	for page := 1; ; page++ {
		u := fmt.Sprintf("%s/payments?page=%d&page_size=%d", s.config.Sync.PaymentsURL, page, pageSize)
		if updatedSince != nil {
			u += "&updated_since=" + updatedSince.Format(time.RFC3339)
		}

		p, err := pkg.RequestWithRetry(ctx, func() (domain.PaymentsPage, error) {
			resp, err := s.hc.Get[domain.PaymentsPage](ctx, u)
			if err != nil {
				return domain.PaymentsPage{}, err
			}
			if resp.StatusCode != http.StatusOK {
				return domain.PaymentsPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
			return resp.Body, nil
		})
		if err != nil {
			s.logger.Error("failed to fetch data from payments-service", "error", err)
			return nil, fmt.Errorf("failed to fetch data from payments-service")
		}

		payments = append(payments, p.Payments...)
		if page >= p.TotalPages {
			return payments, nil
		}
	}
}
