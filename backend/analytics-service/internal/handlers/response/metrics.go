package response

import "time"

type LtvSummaryMetrics struct {
	AvgLtv       int64 `json:"avg_ltv"`
	MedianLtv    int64 `json:"median_ltv"`
	TotalRevenue int64 `json:"total_revenue"`
	Users        int   `json:"users"`
}

type LtvSummaryResponse struct {
	DataAsOf time.Time         `json:"data_as_of"`
	Metrics  LtvSummaryMetrics `json:"metrics"`
}

type RetentionPeriod struct {
	Period         int     `json:"period"`
	Eligible       int     `json:"eligible"`
	Renewed        int     `json:"renewed"`
	Rate           float64 `json:"rate"`
	CumulativeRate float64 `json:"cumulative_rate"`
}

type RetentionResponse struct {
	PlanId     int16             `json:"plan_id"`
	DataAsOf   time.Time         `json:"data_as_of"`
	CohortSize int               `json:"cohort_size"`
	Periods    []RetentionPeriod `json:"periods"`
}

type StatusResponse struct {
	Ready       bool      `json:"ready"`
	LastPoll    time.Time `json:"last_poll"`
	NextPoll    time.Time `json:"next_poll"`
	LastSyncErr string    `json:"last_sync_err,omitempty"`
}
