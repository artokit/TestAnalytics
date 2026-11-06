package handlers

import (
	"errors"
	"net/http"

	_ "analytics-service/internal/handlers/response"

	"analytics-service/internal/domain"
	"analytics-service/internal/handlers/request"
	"analytics-service/internal/services"
	worker "analytics-service/internal/worker"

	"shared/httputil"
)

// GetUserLtvByTelegram godoc
// @Summary      LTV пользователя по telegram_id
// @Description  Пользователь: id, тариф первой подписки, LTV, число платежей и текущий статус подписки. Данные берутся из последнего снимка
// @Produce      json
// @Param        telegram_id path integer true "Telegram ID пользователя"
// @Success      200 {object} response.UserLtvResponse
// @Success      304
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      503 {string} string "снимок ещё не собран"
// @Router       /users/by-telegram/{telegram_id}/ltv [get]
func GetUserLtvByTelegram(readService *services.MetricsReadService, syncer *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telegramId, err := request.GetTelegramId(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		ltv, generatedAt, err := readService.UserLtvByTelegram(r.Context(), telegramId)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSnapshotNotReady):
				snapshotNotReady(w)
				return
			case errors.Is(err, domain.ErrUserNotFound):
				httputil.WriteError(w, err, http.StatusNotFound)
				return
			default:
				httputil.WriteError(w, domain.ErrGetUserLtv, http.StatusInternalServerError)
				return
			}
		}

		writeCached(w, r, generatedAt, syncer.NextPoll(), ltv)
	}
}
