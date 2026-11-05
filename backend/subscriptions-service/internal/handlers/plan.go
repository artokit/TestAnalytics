package handlers

import (
	"net/http"
	"shared/httputil"
	"subscriptions-service/internal/handlers/response"
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
