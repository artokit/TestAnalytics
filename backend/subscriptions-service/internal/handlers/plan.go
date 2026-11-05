package handlers

import (
	"net/http"
	"subscriptions-service/internal/handlers/response"
	"test_analytics/shared/httputil"
)

type IPlanService interface {
	GetPlans() response.PlansResponse
}

func GetPlans(s IPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans := s.GetPlans()
		httputil.WriteJSON(w, plans, http.StatusOK)
	}
}
