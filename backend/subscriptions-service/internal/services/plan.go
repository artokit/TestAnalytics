package services

import (
	"subscriptions-service/internal/handlers/response"
	"subscriptions-service/internal/mock"
)

type PlanService struct {
}

func NewPlanService() *PlanService {
	return &PlanService{}
}

func (s *PlanService) GetPlans() response.PlansResponse {
	return mock.GetPlansMock()
}
