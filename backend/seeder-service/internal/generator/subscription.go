package generator

import (
	"fmt"
	"math/rand"
	"time"
)

// Доли пограничных кейсов (требования задания):
// - ~25% пользователей вообще без единой оплаты;
// - ~5% первых периодов начинаются с неуспешного платежа с повторами;
// - из них ~30% так и не оплачивают (=> тоже "без единой оплаты");
// - ~2.5% успешных периодов рефандятся;
// - ~10% продлений меняют тариф (например месячный -> годовой).
const (
	shareNoPayment     = 0.25
	shareFailedStart   = 0.05
	sharePermanentFail = 0.30
	shareRefund        = 0.025
	sharePlanSwitch    = 0.10

	retryInterval = 24 * time.Hour       // задержка между попытками оплаты
	maxStartDelay = 7 * 24 * time.Hour   // сдвиг первой подписки от регистрации
	maxRenewalGap = 3 * 24 * time.Hour   // пауза между периодами при продлении
	monthDuration = 30 * 24 * time.Hour  // длительность одного периода-месяца
)

// planDef - тариф (цены совместимы с mock subscriptions-service).
type planDef struct {
	id     int16
	months int
	price  int64 // минорные единицы
}

var plans = []planDef{
	{1, 1, 50000},   // месячный
	{2, 3, 135000},  // 3-месячный
	{3, 6, 240000},  // 6-месячный
	{4, 12, 420000}, // годовой
}

// renewalCurves - доля продлений по номеру периода (1-based):
// [0] после 1-го, [1] после 2-го, [2] после 3-го, [3] tail для остальных.
// Оставшиеся продлеваются по tail-константе до конца истории.
var renewalCurves = map[int16][4]float64{
	1: {0.60, 0.75, 0.80, 0.80}, // месячный
	2: {0.70, 0.80, 0.85, 0.85}, // 3-месячный
	3: {0.65, 0.75, 0.80, 0.80}, // 6-месячный
	4: {0.55, 0.70, 0.75, 0.75}, // годовой
}

// initialPlanWeights - популярность тарифов при первой покупке. Сумма = 1.
var initialPlanWeights = [4]float64{0.50, 0.20, 0.18, 0.12}

// Subscription - строка таблицы subscriptions.
type Subscription struct {
	Id          string
	UserId      string
	PlanId      int16
	PeriodStart time.Time
	PeriodEnd   time.Time
	Status      string // active | renewed | expired | refunded
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Payment - строка таблицы payments.
type Payment struct {
	Id             string
	UserId         string
	SubscriptionId string
	Amount         int64
	Currency       string
	Status         string // succeeded | failed | refunded
	PeriodStart    time.Time
	PeriodEnd      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// History - сгенерированная история подписок и платежей.
type History struct {
	Subs     []Subscription
	Payments []Payment
}

// Stats - сводные цифры для лога seeder'а (сколько задевших кейсов).
type Stats struct {
	UsersWithoutPayment int
	FailedPayments      int
	RetriedPayments     int
	Refunds             int
	PlanSwitches        int
	ActiveNow           int
	Renewed             int
	Expired             int
}

// GenerateHistory строит детерминированную по seed историю подписок и платежей
// для пользователей (users уже отсортированы по created_at).
// now - точка отсчёта "сейчас": всё, что позже, в симуляции не возникает.
func GenerateHistory(users []User, seed int64, now time.Time) (*History, *Stats, error) {
	if len(users) == 0 {
		return nil, nil, fmt.Errorf("generator: need users > 0, получено %d", len(users))
	}

	rng := rand.New(rand.NewSource(seed + 7)) //nolint:gosec // seed задан нарочно
	now = now.UTC()

	h := &History{
		Subs:     make([]Subscription, 0),
		Payments: make([]Payment, 0),
	}
	st := &Stats{}

	for _, user := range users {
		if rng.Float64() < shareNoPayment {
			// пользователь без единой оплаты - ни подписок, ни платежей
			st.UsersWithoutPayment++
			continue
		}

		plan := pickInitialPlan(rng)
		start := user.CreatedAt.Add(time.Duration(rng.Int63n(int64(maxStartDelay))))
		if !start.Before(now) {
			// слишком свежая регистрация - до первой оплаты ещё не дошёл
			st.UsersWithoutPayment++
			continue
		}

		for periodNo := 1; ; periodNo++ {
			def := plans[plan-1]
			periodStart := start
			periodEnd := periodStart.Add(time.Duration(def.months) * monthDuration)

			// --- оплата периода (возможны попытки с повторами) ---
			created := periodStart
			if periodNo == 1 && rng.Float64() < shareFailedStart {
				fails := 1 + rng.Intn(2) // 1-2 неуспешных попытки
				for i := 1; i <= fails; i++ {
					attemptAt := periodStart.Add(time.Duration(i-1) * retryInterval)
					h.addPayment(rng, user.Id, newUUIDv7String(attemptAt.UnixMilli(), rng),
						def.price, "failed", periodStart, periodEnd, attemptAt)
					st.FailedPayments++
				}

				if rng.Float64() < sharePermanentFail {
					// так и не оплатил: строк подписки нет,
					// попытки остались в payments с синтетическим subscription_id
					st.UsersWithoutPayment++
					break
				}

				// успех на следующей попытке
				created = periodStart.Add(time.Duration(fails) * retryInterval)
				st.RetriedPayments++
			}

			subsId := newUUIDv7String(created.UnixMilli(), rng)
			h.addSub(subsId, user.Id, plan, periodStart, periodEnd, created)
			h.addPayment(rng, user.Id, subsId, def.price, "succeeded", periodStart, periodEnd, created)

			// --- рефанд ---
			if rng.Float64() < shareRefund {
				h.markLastPaymentRefunded(rng, now)
				st.Refunds++
				h.setLastSubStatus("canceled")
				break
			}

			if periodEnd.After(now) {
				// период ещё не закончился
				h.setLastSubStatus("active")
				st.ActiveNow++
				break
			}

			// --- решение о продлении ---
			if rng.Float64() >= renewalProb(plan, periodNo) {
				// период закончился без продления - точка оттока
				h.setLastSubStatus("expired")
				st.Expired++
				break
			}

			// продлён период: следующий период порождает новую строку,
			// продление выводится из связки (next period_start >= period_end)
			h.setLastSubStatus("expired")
			st.Renewed++

			if rng.Float64() < sharePlanSwitch {
				plan = switchPlan(plan, rng)
				st.PlanSwitches++
			}

			start = periodEnd.Add(time.Duration(rng.Int63n(int64(maxRenewalGap))))
		}
	}

	return h, st, nil
}

// addPayment добавляет строку платежа. subscription_id передаётся явно:
// для группы неудачных попыток это синтетический UUID без строки в subscriptions.
func (h *History) addPayment(rng *rand.Rand, userId string, subsId string, amount int64, status string, periodStart, periodEnd, createdAt time.Time) {
	h.Payments = append(h.Payments, Payment{
		Id:             newUUIDv7String(createdAt.UnixMilli(), rng),
		UserId:         userId,
		SubscriptionId: subsId,
		Amount:         amount,
		Currency:       "RUB",
		Status:         status,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
}

func (h *History) addSub(id, userId string, plan int16, periodStart, periodEnd, createdAt time.Time) {
	h.Subs = append(h.Subs, Subscription{
		Id:          id,
		UserId:      userId,
		PlanId:      plan,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      "", // выставляется ниже: active / churned / refunded
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

func (h *History) setLastSubStatus(status string) {
	if len(h.Subs) > 0 {
		h.Subs[len(h.Subs)-1].Status = status
	}
}

// refundDelayMax: рефанд случается позже оплаты (0..N дней) -
// updated_at refunded-платежа отражает это "событие смены статуса".
const refundDelayMax = 14 * 24 * time.Hour

func (h *History) markLastPaymentRefunded(rng *rand.Rand, now time.Time) {
	for i := len(h.Payments) - 1; i >= 0; i-- {
		if h.Payments[i].Status == "succeeded" {
			h.Payments[i].Status = "refunded"
			updatedAt := h.Payments[i].CreatedAt.Add(time.Duration(rng.Int63n(int64(refundDelayMax))))
			if updatedAt.After(now) {
				updatedAt = now
			}
			h.Payments[i].UpdatedAt = updatedAt
			return
		}
	}
}

// renewalProb - доля продления для тарифа по номеру периода (1-based).
func renewalProb(plan int16, periodNo int) float64 {
	curve := renewalCurves[plan]
	switch {
	case periodNo <= 1:
		return curve[0]
	case periodNo == 2:
		return curve[1]
	case periodNo == 3:
		return curve[2]
	default:
		return curve[3]
	}
}

func pickInitialPlan(rng *rand.Rand) int16 {
	r := rng.Float64()
	cum := 0.0
	for i, w := range initialPlanWeights {
		cum += w
		if r < cum {
			return int16(i + 1)
		}
	}
	return 1
}

// switchPlan - смена тарифа при продлении (не такой же).
func switchPlan(cur int16, rng *rand.Rand) int16 {
	next := int16(1 + rng.Intn(len(plans)-1)) // 1..3
	if next >= cur {
		next++
	}
	return next
}
