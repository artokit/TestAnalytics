package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"subscriptions-service/internal/handlers/request"
	"subscriptions-service/internal/handlers/response"
	"test_analytics/shared/httputil"
)

type ISubscriptionService interface {
	GetSubscriptions(ctx context.Context, filter *request.SubscriptionFilter) (response.SubscriptionsResponsePaged, error)
}

// GetSubscriptions godoc
// @Summary      Список подписок
// @Description  Возвращает подписки постранично, сортировка по created_at (новые сначала)
// @Tags         subscriptions
// @Produce      json
// @Param        page       query integer false "Номер страницы (минимум 1)"    minimum(1) default(1)
// @Param        page_size  query integer false "Размер страницы"             minimum(1) default(50)
// @Success      200 {object} response.SubscriptionsResponsePaged
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /subscriptions [get]
func GetSubscriptions(service ISubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetSubscriptionFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		res, err := service.GetSubscriptions(r.Context(), filter)
		if err != nil {
			httputil.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}
