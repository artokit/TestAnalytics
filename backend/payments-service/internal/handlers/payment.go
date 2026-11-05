package handlers

import (
	"context"
	"net/http"
	"payments-service/internal/handlers/request"
	"payments-service/internal/handlers/response"
	"test_analytics/shared/httputil"
)

type IPaymentService interface {
	GetPayments(ctx context.Context, filter *request.PaymentFilter) (response.PaymentsResponsePaged, error)
}

func GetPayments(service IPaymentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetPaymentFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		res, err := service.GetPayments(r.Context(), filter)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		httputil.WriteJSON(w, res, http.StatusOK)
	}
}
