package generator

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"seeder-service/internal/models"

	"github.com/google/uuid"
)

const (
	telegramIDMin = 100_000_000
	telegramIDMax = 7_000_000_000
)

const HistoryMonths = 24

var utmLinkPool = []string{
	"utm_source=telegram&utm_medium=post&utm_campaign=spring_sale",
	"utm_source=telegram&utm_medium=bot&utm_campaign=renewal_mar",
	"utm_source=telegram&utm_medium=cpc&utm_campaign=chai_launch",
	"utm_source=telegram&utm_medium=post&utm_campaign=autumn_dis",
	"utm_source=partner&utm_medium=referral&utm_campaign=vpn_blog",
	"utm_source=telegram&utm_medium=post&utm_campaign=black_friday",
	"utm_source=tg_ads&utm_medium=cpc&utm_campaign=retarget_q4",
	"utm_source=telegram&utm_medium=inf&utm_campaign=tg_channels",
	"utm_source=influencer&utm_medium=post&utm_campaign=streamers",
	"utm_source=docs&utm_medium=content&utm_campaign=seo_q1",
}

const shareOfEmptyUtm = 0.2

func GenerateUsers(n int, seed int64, now time.Time, historyMonths int) ([]models.User, error) {
	if n <= 0 {
		return nil, fmt.Errorf("generator: need n > 0, получено %d", n)
	}

	rng := rand.New(rand.NewSource(seed))
	now = now.UTC()

	telegramIDs, err := generateTelegramIDs(n, rng)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, n)
	if historyMonths <= 0 {
		historyMonths = HistoryMonths
	}
	historyStart := now.AddDate(0, -historyMonths, 0)
	historySpan := now.Sub(historyStart)

	for i := range n {
		createdAt := historyStart.Add(time.Duration(rng.Int63n(int64(historySpan)))).
			Truncate(time.Millisecond)

		users = append(users, models.User{
			Id:         uuid.NewString(),
			TelegramId: telegramIDs[i],
			CreatedAt:  createdAt,
			UtmLink:    pickUtmLink(rng),
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})

	return users, nil
}

func generateTelegramIDs(n int, rng *rand.Rand) ([]int64, error) {
	ids := make([]int64, 0, n)
	seen := make(map[int64]struct{}, n)
	span := uint64(telegramIDMax - telegramIDMin + 1)

	for len(ids) < n {
		id := telegramIDMin + int64(rng.Uint64()%span)
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	return ids, nil
}

func pickUtmLink(rng *rand.Rand) string {
	if rng.Float64() < shareOfEmptyUtm {
		return ""
	}
	return utmLinkPool[rng.Intn(len(utmLinkPool))]
}
