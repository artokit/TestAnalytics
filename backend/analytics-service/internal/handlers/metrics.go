package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"analytics-service/internal/handlers/request"
	"analytics-service/internal/handlers/response"
	"analytics-service/internal/services"
	worker "analytics-service/internal/worker"

	"shared/httputil"
)

const maxCacheControlAgeSeconds = 600

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
func GetLtvSummary(readService *services.MetricsReadService, syncer *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetMetricsFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		summary, err := readService.Summary(r.Context(), filter.From, filter.To)
		if err != nil {
			if errors.Is(err, services.ErrSnapshotNotReady) {
				snapshotNotReady(w)
				return
			}
			httputil.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		writeCached(w, r, summary.DataAsOf, syncer.NextPoll(), response.LtvSummaryResponse{
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
func GetRetention(readService *services.MetricsReadService, syncer *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetRetentionFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		plan, err := readService.Retention(r.Context(), filter.PlanId, filter.From, filter.To, filter.MaxPeriods)
		if err != nil {
			if errors.Is(err, services.ErrSnapshotNotReady) {
				snapshotNotReady(w)
				return
			}
			httputil.WriteError(w, err, http.StatusInternalServerError)
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

		writeCached(w, r, plan.DataAsOf, syncer.NextPoll(), response.RetentionResponse{
			PlanId:     plan.PlanId,
			DataAsOf:   plan.DataAsOf,
			CohortSize: plan.CohortSize,
			Periods:    respPeriods,
		})
	}
}

// GetLtvByPlan godoc
// @Summary      LTV по тарифу первой подписки
// @Description  Средний LTV, число пользователей и выручка по тарифам первой подписки когорты. from/to фильтруют когорту по дате первой подписки (RFC3339)
// @Produce      json
// @Param        from query string false "Начало когорты (RFC3339)"
// @Param        to   query string false "Конец когорты (RFC3339)"
// @Success      200 {object} response.LtvByPlanResponse
// @Success      304
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      503 {string} string "снимок ещё не собран"
// @Router       /metrics/ltv/by-plan [get]
func GetLtvByPlan(readService *services.MetricsReadService, syncer *worker.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetMetricsFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		byPlan, err := readService.LtvByPlan(r.Context(), filter.From, filter.To)
		if err != nil {
			if errors.Is(err, services.ErrSnapshotNotReady) {
				snapshotNotReady(w)
				return
			}
			httputil.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		plans := make([]response.PlanLtv, 0, len(byPlan.Plans))
		for _, plan := range byPlan.Plans {
			plans = append(plans, response.PlanLtv{
				PlanId:       plan.PlanId,
				Users:        plan.Users,
				AvgLtv:       plan.AvgLtv,
				TotalRevenue: plan.TotalRevenue,
			})
		}

		writeCached(w, r, byPlan.DataAsOf, syncer.NextPoll(), response.LtvByPlanResponse{
			DataAsOf: byPlan.DataAsOf,
			Plans:    plans,
		})
	}
}

func snapshotNotReady(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "60")
	httputil.WriteJSON(w, map[string]string{
		"error": "data snapshot is not ready yet, try again later",
	}, http.StatusServiceUnavailable)
}

func writeCached(w http.ResponseWriter, r *http.Request, dataAsOf time.Time, nextPoll time.Time, body any) {
	etag := `"` + dataAsOf.UTC().Format(time.RFC3339Nano) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", dataAsOf.UTC().Format(http.TimeFormat))

	if !nextPoll.IsZero() {
		maxAge := int(time.Until(nextPoll).Seconds())
		if maxAge < 0 {
			maxAge = 0
		}
		if maxAge > maxCacheControlAgeSeconds {
			maxAge = maxCacheControlAgeSeconds
		}
		w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(maxAge))
	}

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	httputil.WriteJSON(w, body, http.StatusOK)
}
