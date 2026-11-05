package generator

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"seeder-service/internal/models"
)

func addSub(h *models.History, id, userId string, plan int16, periodStart, periodEnd, createdAt time.Time) {
	h.Subs = append(h.Subs, models.Subscription{
		Id:          id,
		UserId:      userId,
		PlanId:      plan,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      "",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

func addPayment(h *models.History, userId, subsId string, price int64, status string, periodStart, periodEnd, createdAt time.Time) {
	h.Payments = append(h.Payments, models.Payment{
		Id:             uuid.NewString(),
		UserId:         userId,
		SubscriptionId: subsId,
		Amount:         price,
		Currency:       "RUB",
		Status:         status,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
}

func setLastSubStatus(h *models.History, status string) {
	if len(h.Subs) > 0 {
		h.Subs[len(h.Subs)-1].Status = status
	}
}

func markLastPaymentRefunded(h *models.History, now time.Time, rng *rand.Rand) {
	for i := len(h.Payments) - 1; i >= 0; i-- {
		if h.Payments[i].Status != "succeeded" {
			continue
		}

		h.Payments[i].Status = "refunded"
		h.Payments[i].UpdatedAt = refundMoment(h.Payments[i].CreatedAt, now, rng)
		return
	}
}

func refundMoment(paidAt, now time.Time, rng *rand.Rand) time.Time {
	moment := paidAt.Add(time.Duration(rng.Int63n(int64(refundDelay))))
	if moment.After(now) {
		return now
	}
	return moment
}
