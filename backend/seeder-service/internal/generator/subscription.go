package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"seeder-service/internal/models"
)

const (
	shareNoPayment     = 0.25
	shareFailedStart   = 0.05
	sharePermanentFail = 0.30
	shareRefund        = 0.025
	sharePlanSwitch    = 0.10

	retryInterval = 24 * time.Hour
	maxStartDelay = 7 * 24 * time.Hour
	maxRenewalGap = 3 * 24 * time.Hour
	monthDuration = 30 * 24 * time.Hour
	refundDelay   = 14 * 24 * time.Hour
)

func GenerateHistory(users []models.User, seed int64, now time.Time) (*models.History, error) {
	if len(users) == 0 {
		return nil, fmt.Errorf("generator: нужно users > 0, получено %d", len(users))
	}

	rng := rand.New(rand.NewSource(seed + 7))
	h := models.NewHistory()

	for _, user := range users {
		generateUserHistory(h, user, now.UTC(), rng)
	}

	return h, nil
}

func generateUserHistory(h *models.History, user models.User, now time.Time, rng *rand.Rand) {
	if userSkipsPayment(rng) {
		return
	}

	plan := pickInitialPlan(rng)
	start := firstPeriodStart(user, rng)
	if start.After(now) {
		return
	}

	for periodNo := 1; ; periodNo++ {
		periodStart := start
		periodEnd := periodStart.Add(planDuration(plan))

		createdAt := payForPeriod(h, user.Id, plan, periodNo, periodStart, periodEnd, now, rng)
		if createdAt.IsZero() {
			return
		}

		subsId := uuid.NewString()
		addSub(h, subsId, user.Id, plan, periodStart, periodEnd, createdAt)
		addPayment(h, user.Id, subsId, planPrice(plan), "succeeded", periodStart, periodEnd, createdAt)

		if refunded(h, now, rng) {
			setLastSubStatus(h, "canceled")
			return
		}

		if periodEnd.After(now) {
			setLastSubStatus(h, "active")
			return
		}

		if !renews(plan, periodNo, rng) {
			setLastSubStatus(h, "expired")
			return
		}

		setLastSubStatus(h, "expired")

		if rng.Float64() < sharePlanSwitch {
			plan = switchPlan(plan, rng)
		}

		start = periodEnd.Add(time.Duration(rng.Int63n(int64(maxRenewalGap))))
	}
}

func payForPeriod(
	h *models.History,
	userId string,
	plan int16,
	periodNo int,
	periodStart, periodEnd, now time.Time,
	rng *rand.Rand,
) time.Time {
	if periodNo == 1 && rng.Float64() < shareFailedStart {
		return failedStart(h, userId, plan, periodStart, periodEnd, now, rng)
	}
	return periodStart
}

func failedStart(
	h *models.History,
	userId string,
	plan int16,
	periodStart, periodEnd, now time.Time,
	rng *rand.Rand,
) time.Time {
	price := planPrice(plan)
	fails := 1 + rng.Intn(2)

	for i := 1; i <= fails; i++ {
		attemptAt := periodStart.Add(time.Duration(i-1) * retryInterval)
		addPayment(h, userId, uuid.NewString(), price, "failed", periodStart, periodEnd, attemptAt)
	}

	if rng.Float64() < sharePermanentFail {
		return time.Time{}
	}

	return periodStart.Add(time.Duration(fails) * retryInterval)
}

func refunded(h *models.History, now time.Time, rng *rand.Rand) bool {
	if rng.Float64() >= shareRefund {
		return false
	}

	markLastPaymentRefunded(h, now, rng)
	return true
}

func renews(plan int16, periodNo int, rng *rand.Rand) bool {
	return rng.Float64() < renewalProb(plan, periodNo)
}

func userSkipsPayment(rng *rand.Rand) bool {
	return rng.Float64() < shareNoPayment
}

func firstPeriodStart(user models.User, rng *rand.Rand) time.Time {
	return user.CreatedAt.Add(time.Duration(rng.Int63n(int64(maxStartDelay))))
}
