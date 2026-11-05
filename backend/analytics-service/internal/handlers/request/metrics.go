package request

import "time"

//since metrics короков are-coming; contract-ready парсеры query-параметров.

type MetricsFilter struct {
	// From / To - границы когорты по дате ПЕРВОЙ подписки (RFC3339).
	From *time.Time
	To   *time.Time
}

// PlanId - в примере ТЗ строка ("monthly"), храним контрактно как строку.
type RetentionFilter struct {
	MetricsFilter
	PlanId     string
	MaxPeriods int
}
