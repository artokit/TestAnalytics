package response

type UserLtvResponse struct {
	UserId        string `json:"user_id"`
	TelegramId    int64  `json:"telegram_id"`
	FirstPlanId   int16  `json:"first_plan_id"`
	Ltv           int64  `json:"ltv"`
	PaymentsCount int    `json:"payments_count"`
	CurrentStatus string `json:"current_status"`
}
