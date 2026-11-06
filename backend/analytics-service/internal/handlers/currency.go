package handlers

import (
	"net/http"

	"analytics-service/internal/services"

	"shared/httputil"
)

func GetCurrency(svc *services.CurrencyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := svc.Prices(r.Context())
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		httputil.WriteJSON(w, res, http.StatusOK)
	}
}
