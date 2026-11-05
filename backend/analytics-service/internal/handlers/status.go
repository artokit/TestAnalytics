package handlers

import (
	"net/http"

	"analytics-service/internal/handlers/response"
	worker "analytics-service/internal/worker"

	"shared/httputil"
)

// GetStatus godoc
// @Summary      Статус синхронизации
// @Description  Время последнего успешного опроса источников, время следующего опроса и последняя ошибка синхронизации, если была
// @Tags         system
// @Produce      json
// @Success      200 {object} response.StatusResponse
// @Router       /status [get]
func GetStatus(worker *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lastPoll, nextPoll, lastErr := worker.Status()

		httputil.WriteJSON(w, response.StatusResponse{
			Ready:       worker.Ready(),
			LastPoll:    lastPoll,
			NextPoll:    nextPoll,
			LastSyncErr: lastErr,
		}, http.StatusOK)
	}
}

// Healthz godoc
// @Summary      Живость сервиса
// @Description  200 если процесс отвечает
// @Tags         system
// @Produce      json
// @Success      200 {object} response.HealthzResponse
// @Router       /healthz [get]
func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, response.HealthzResponse{Status: "ok"}, http.StatusOK)
	}
}
