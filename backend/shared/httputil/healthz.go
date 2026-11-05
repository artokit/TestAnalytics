package httputil

import "net/http"

func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, map[string]string{"status": "ok"}, http.StatusOK)
	}
}
