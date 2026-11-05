package handlers

import (
	"net/http"
	"test_analytics/shared/httputil"
	"subscriptions-service/internal/handlers/response"
)

type IPlanService interface {
	GetPlans() response.PlansResponse
}

// GetPlans godoc
// @Summary      Список тарифных планов
// @Description  Возвращает доступные тарифные планы подписок
// @Tags         plans
// @Produce      json
// @Success      200 {object} response.PlansResponse
// @Router       /plans [get]
func GetPlans(s IPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans := s.GetPlans()
		httputil.WriteJSON(w, plans, http.StatusOK)
	}
}
