package handlers

import (
	"net/http"
	"sort"

	"analytics-service/internal/domain"
	worker "analytics-service/internal/worker"

	"shared/httputil"
)

// GetPlans godoc
// @Summary      Список тарифных планов
// @Description  Планы из снимка данных воркера (источник subscriptions-service)
// @Produce      json
// @Success      200 {object} domain.PlansResponse
// @Router       /plans [get]
func GetPlans(worker *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worker.Read(func(_ map[string]domain.User, _ map[string]domain.Payment, _ map[string]domain.Subscription, plans map[int]domain.Plan) {
			items := make([]domain.Plan, 0, len(plans))
			for _, plan := range plans {
				items = append(items, plan)
			}
			sort.Slice(items, func(i, j int) bool { return items[i].Id < items[j].Id })
			httputil.WriteJSON(w, domain.PlansResponse{Items: items}, http.StatusOK)
		})
	}
}
