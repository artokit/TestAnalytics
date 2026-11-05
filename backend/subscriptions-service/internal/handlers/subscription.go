package handlers

import (
	"context"
	"net/http"
	"subscriptions-service/internal/handlers/request"
	"subscriptions-service/internal/handlers/response"
	"test_analytics/shared/httputil"
)

type ISubscriptionService interface {
	GetSubscriptions(ctx context.Context, filter *request.SubscriptionFilter) (response.SubscriptionsResponsePaged, error)
}

func GetSubscriptions(service ISubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetSubscriptionFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		res, err := service.GetSubscriptions(r.Context(), filter)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		httputil.WriteJSON(w, res, http.StatusOK)
	}
}
