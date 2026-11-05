package response

type PlansResponse struct {
	Items []PlanResponse `json:"items"`
}

type PlanResponse struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	PeriodMonths int    `json:"period_months"`
	Price        int64  `json:"price"`
	Currency     string `json:"currency"`
}
