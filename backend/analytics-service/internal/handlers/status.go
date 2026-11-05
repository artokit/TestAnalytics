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
		lastPoll, nextPoll, ready, lastSyncErr := worker.Status()

		httputil.WriteJSON(w, response.StatusResponse{
			Ready:       ready,
			LastPoll:    lastPoll,
			NextPoll:    nextPoll,
			LastSyncErr: lastSyncErr,
		}, http.StatusOK)
	}
}
