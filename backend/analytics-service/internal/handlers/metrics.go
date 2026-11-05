package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"analytics-service/internal/handlers/response"
	"analytics-service/internal/services"

	"shared/httputil"
)

const defaultMaxPeriods = 12

func snapshotNotReady(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "60")
	httputil.WriteJSON(w, map[string]string{
		"error": "data snapshot is not ready yet, try again later",
	}, http.StatusServiceUnavailable)
}

func parseFromTo(r *http.Request) (*time.Time, *time.Time, error) {
	query := r.URL.Query()

	var from, to *time.Time

	if raw := query.Get("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, nil, fmt.Errorf("неверно передан параметр from, expected RFC3339")
		}
		from = &t
	}

	if raw := query.Get("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, nil, fmt.Errorf("неверно передан параметр to, expected RFC3339")
		}
		to = &t
	}

	return from, to, nil
}

func writeCached(w http.ResponseWriter, r *http.Request, dataAsOf time.Time, body any) {
	etag := `"` + dataAsOf.Format(time.RFC3339Nano) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", dataAsOf.UTC().Format(http.TimeFormat))
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	httputil.WriteJSON(w, body, http.StatusOK)
}

// GetLtvSummary godoc
// @Summary      Общий LTV по всем пользователям
// @Description  Среднее, медиана, общая выручка и число пользователей когорты. from/to фильтруют когорту по дате первой подписки (RFC3339)
// @Produce      json
// @Param        from query string false "Начало когорты (RFC3339)"
// @Param        to   query string false "Конец когорты (RFC3339)"
// @Success      200 {object} response.LtvSummaryResponse
// @Success      304
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      503 {string} string "снимок ещё не собран"
// @Router       /metrics/ltv/summary [get]
func GetLtvSummary(readService *services.MetricsReadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		from, to, err := parseFromTo(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		summary, found, err := readService.Summary(r.Context(), from, to)
		if err != nil {
			httputil.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		if !found {
			snapshotNotReady(w)
			return
		}

		writeCached(w, r, summary.DataAsOf, response.LtvSummaryResponse{
			DataAsOf: summary.DataAsOf,
			Metrics: response.LtvSummaryMetrics{
				AvgLtv:       summary.AvgLtv,
				MedianLtv:    summary.MedianLtv,
				TotalRevenue: summary.TotalRevenue,
				Users:        summary.Users,
			},
		})
	}
}

// GetRetention godoc
// @Summary      Процент продлений по периодам для тарифа
// @Description  Условные и кумулятивные доли продлений по периодам (когорта - пользователи с первой подпиской на указанный тариф). from/to фильтруют когорту по дате первой подписки
// @Produce      json
// @Param        plan_id     query integer true  "Идентификатор тарифа"
// @Param        max_periods query integer false "Максимальное число периодов (по умолчанию 12)" minimum(1) default(12)
// @Param        from        query string  false "Начало когорты (RFC3339)"
// @Param        to          query string  false "Конец когорты (RFC3339)"
// @Success      200 {object} response.RetentionResponse
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse "нет данных для тарифа"
// @Failure      503 {string} string "снимок ещё не собран"
// @Failure      304
// @Router       /metrics/retention [get]
func GetRetention(readService *services.MetricsReadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		planId, err := strconv.Atoi(query.Get("plan_id"))
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		maxPeriods := defaultMaxPeriods
		if raw := query.Get("max_periods"); raw != "" {
			maxPeriods, err = strconv.Atoi(raw)
			if err != nil {
				httputil.WriteError(w, err, http.StatusBadRequest)
				return
			}
			if maxPeriods < 1 {
				httputil.WriteError(w, fmt.Errorf("max_periods must be >= 1"), http.StatusBadRequest)
				return
			}
		}

		from, to, err := parseFromTo(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		plan, found, err := readService.Retention(r.Context(), int16(planId), from, to, maxPeriods)
		if err != nil {
			httputil.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		if !found {
			snapshotNotReady(w)
			return
		}

		if plan.CohortSize == 0 && len(plan.Periods) == 0 {
			httputil.WriteJSON(w, map[string]string{
				"error": "no retention data for requested plan/filters",
			}, http.StatusNotFound)
			return
		}

		respPeriods := make([]response.RetentionPeriod, 0, len(plan.Periods))
		for _, p := range plan.Periods {
			respPeriods = append(respPeriods, response.RetentionPeriod{
				Period:         p.Period,
				Eligible:       p.Eligible,
				Renewed:        p.Renewed,
				Rate:           p.Rate,
				CumulativeRate: p.CumulativeRate,
			})
		}

		writeCached(w, r, plan.DataAsOf, response.RetentionResponse{
			PlanId:     plan.PlanId,
			DataAsOf:   plan.DataAsOf,
			CohortSize: plan.CohortSize,
			Periods:    respPeriods,
		})
	}
}

// GetLtvByPlan godoc
// @Summary      LTV по тарифу первой подписки
// @Description  Стаб
// @Produce      json
// @Failure      503 {string} string "снимок ещё не собран"
// @Router       /metrics/ltv/by-plan [get]
func GetLtvByPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshotNotReady(w)
	}
}
