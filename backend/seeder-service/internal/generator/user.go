package generator

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

const (
	// Правдоподобный диапазон для пользовательских Telegram ID
	// (не боты, не 1,2,3). Нижняя граница - старые аккаунты,
	// верхняя - примерно текущая верхушка нумерации.
	telegramIDMin = 100_000_000
	telegramIDMax = 7_000_000_000
)

// HistoryMonths - глубина истории регистраций:
// пользователи появляются равномерно за последние N месяцев.
const HistoryMonths = 24

// User - строка таблицы users (миграция users-service, неизменная).
type User struct {
	Id         string
	TelegramId int64
	CreatedAt  time.Time
	UtmLink    string
}

// utmLinkPool - правдоподобные UTM-ссылки трафика. Пул фиксированный
// для воспроизводимости; впоследствии он же станет основой для
// разбивки метрик по utm_link.
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

// shareOfEmptyUtm - доля регистраций без UTM ("прямой" трафик).
const shareOfEmptyUtm = 0.2

// GenerateUsers возвращает n пользователей, детерминированно по seed.
// now - точка отсчёта истории (передаётся явно для тестируемости).
// historyMonths - глубина истории регистраций (18-24 месяца); <=0 -> значение по умолчанию.
// Результат отсортирован по created_at asc - вместе с UUIDv7 это даёт
// почти монотонную вставку идентификаторов и отсутствие расщепления
// страниц PK-индекса.
func GenerateUsers(n int, seed int64, now time.Time, historyMonths int) ([]User, error) {
	if n <= 0 {
		return nil, fmt.Errorf("generator: need n > 0, получено %d", n)
	}

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // seed задан нарочно
	now = now.UTC()

	telegramIDs, err := generateTelegramIDs(n, rng)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, n)
	if historyMonths <= 0 {
		historyMonths = HistoryMonths
	}
	historyStart := now.AddDate(0, -historyMonths, 0)
	historySpan := now.Sub(historyStart)

	for i := 0; i < n; i++ {
		createdAt := historyStart.Add(time.Duration(rng.Int63n(int64(historySpan)))).
			Truncate(time.Millisecond)

		users = append(users, User{
			Id:         newUUIDv7String(createdAt.UnixMilli(), rng),
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
