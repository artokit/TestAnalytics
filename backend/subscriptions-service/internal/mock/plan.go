package mock

import "subscriptions-service/internal/handlers/response"

func GetPlansMock() response.PlansResponse {
	return response.PlansResponse{
		Items: []response.PlanResponse{
			{
				Id:           1,
				Name:         "Месячная подписка",
				PeriodMonths: 1,
				Price:        50000,
				Currency:     "RUB",
			},
			{
				Id:           2,
				Name:         "Подписка на 3 месяца",
				PeriodMonths: 3,
				Price:        135000,
				Currency:     "RUB",
			},
			{
				Id:           3,
				Name:         "Подписка на 6 месяцев",
				PeriodMonths: 6,
				Price:        240000,
				Currency:     "RUB",
			},
			{
				Id:           4,
				Name:         "Годовая подписка",
				PeriodMonths: 12,
				Price:        420000,
				Currency:     "RUB",
			},
		},
	}
}
