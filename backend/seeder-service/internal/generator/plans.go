package generator

import (
	"math/rand"
	"time"
)

type planDef struct {
	id     int16
	months int
	price  int64
}

var plans = []planDef{
	{1, 1, 50000},
	{2, 3, 135000},
	{3, 6, 240000},
	{4, 12, 420000},
}

var renewalCurves = map[int16][4]float64{
	1: {0.60, 0.75, 0.80, 0.80},
	2: {0.70, 0.80, 0.85, 0.85},
	3: {0.65, 0.75, 0.80, 0.80},
	4: {0.55, 0.70, 0.75, 0.75},
}

var initialPlanWeights = [4]float64{0.50, 0.20, 0.18, 0.12}

func planDuration(plan int16) time.Duration {
	return time.Duration(plans[plan-1].months) * monthDuration
}

func planPrice(plan int16) int64 {
	return plans[plan-1].price
}

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
	roll := rng.Float64()
	cum := 0.0
	for i, weight := range initialPlanWeights {
		cum += weight
		if roll < cum {
			return int16(i + 1)
		}
	}
	return 1
}

func switchPlan(cur int16, rng *rand.Rand) int16 {
	next := int16(1 + rng.Intn(len(plans)-1))
	if next >= cur {
		next++
	}
	return next
}
