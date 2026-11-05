package domain

type Plan struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	PeriodMonths int    `json:"period_months"`
	Price        int64  `json:"price"`
	Currency     string `json:"currency"`
}

type PlansResponse struct {
	Items []Plan `json:"items"`
}
