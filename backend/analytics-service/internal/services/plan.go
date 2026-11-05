package services

import (
	"context"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/pkg"
)

type PlanService struct {
	hc     *httpclient.HttpClient
	config *config.Config
}

func NewPlanService(hc *httpclient.HttpClient, cfg *config.Config) *PlanService {
	return &PlanService{
		hc:     hc,
		config: cfg,
	}
}

func (s *PlanService) GetPlans(ctx context.Context, url string) (httpclient.Response[domain.PlansResponse], error) {
	return s.hc.Get[domain.PlansResponse](ctx, url)
}
