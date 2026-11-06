package services

import (
	"context"
	"log/slog"
	"net/http"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/internal/handlers/response"
	"analytics-service/pkg"
)

type CurrencyService struct {
	hc     *pkg.HttpClient
	logger *slog.Logger
	config *config.Config
}

func NewCurrencyService(hc *pkg.HttpClient, logger *slog.Logger, cfg *config.Config) *CurrencyService {
	return &CurrencyService{
		hc:     hc,
		logger: logger,
		config: cfg,
	}
}

func (s *CurrencyService) Prices(ctx context.Context) (response.CurrencyResponse, error) {
	resp, err := s.hc.Get[response.CoinGeckoRates](ctx, s.config.Currency.PricesURL)
	if err != nil {
		s.logger.Error("failed to fetch currency prices", "error", err)
		return response.CurrencyResponse{}, domain.ErrGetCurrencyPrices
	}
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("failed to fetch currency prices", "status", resp.StatusCode)
		return response.CurrencyResponse{}, domain.ErrGetCurrencyPrices
	}

	return response.CurrencyResponse{
		Base: "RUB",
		Prices: map[string]float64{
			"USDT": resp.Body.Tether.Rub,
			"SOL":  resp.Body.Solana.Rub,
		},
	}, nil
}
