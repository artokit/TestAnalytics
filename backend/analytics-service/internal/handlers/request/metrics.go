package request

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"shared/httputil"
)

const defaultMaxPeriods = 12

type MetricsFilter struct {
	From *time.Time
	To   *time.Time
}

type RetentionFilter struct {
	MetricsFilter
	PlanId     int16
	MaxPeriods int
}

func GetMetricsFilter(r *http.Request) (MetricsFilter, error) {
	from, err := httputil.ReadTimeQuery(r, "from")
	if err != nil {
		return MetricsFilter{}, fmt.Errorf("неверно передан параметр from, expected RFC3339")
	}

	to, err := httputil.ReadTimeQuery(r, "to")
	if err != nil {
		return MetricsFilter{}, fmt.Errorf("неверно передан параметр to, expected RFC3339")
	}

	return MetricsFilter{From: from, To: to}, nil
}

func GetRetentionFilter(r *http.Request) (RetentionFilter, error) {
	metrics, err := GetMetricsFilter(r)
	if err != nil {
		return RetentionFilter{}, err
	}

	planId, err := strconv.Atoi(r.URL.Query().Get("plan_id"))
	if err != nil {
		return RetentionFilter{}, fmt.Errorf("неверно передан параметр plan_id, expected integer")
	}

	maxPeriods := defaultMaxPeriods
	if raw := r.URL.Query().Get("max_periods"); raw != "" {
		maxPeriods, err = strconv.Atoi(raw)
		if err != nil {
			return RetentionFilter{}, fmt.Errorf("неверно передан параметр max_periods, expected integer")
		}
		if maxPeriods < 1 {
			return RetentionFilter{}, fmt.Errorf("max_periods must be >= 1")
		}
	}

	return RetentionFilter{
		MetricsFilter: metrics,
		PlanId:        int16(planId),
		MaxPeriods:    maxPeriods,
	}, nil
}
