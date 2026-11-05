package generator

import (
	"encoding/binary"
	"math/rand"
)

// newUUIDv7 формирует UUID v7 из unix-времени (мс) и seeded-источника
// случайных бит r. Детерминирован при том же (ts, r).
//
// Формат (RFC 9562):
//   - байты 0-5: unix_ts_ms (48 бит) -> сортировка по времени,
//     вставки отсортированных строк идут в "хвост" B-tree индекса;
//   - байт 6, старшая тетрада: версия 0x7;
//   - байт 6 младшая тетрада + байт 7: rand_a (12 бит);
//   - байт 8, старшие два бита: вариант 10xx;
//   - остальное: rand_b (62 бита).
func newUUIDv7(timestampUnixMilli int64, r *rand.Rand) [16]byte {
	var b [16]byte

	binary.BigEndian.PutUint64(b[:8], uint64(timestampUnixMilli))
	b[6] = 0x70 | byte(r.Uint64()&0x0f)
	b[7] = byte(r.Uint64() >> 8)
	b[8] = 0x80 | byte(r.Uint64()&0x3f)
	for i := 9; i < 16; i++ {
		b[i] = byte(r.Uint64() >> 8)
	}

	return b
}

const hexDigits = "0123456789abcdef"

func formatUUID(b [16]byte) string {
	// канонический вывод 8-4-4-4-12
	buf := make([]byte, 0, 36)
	for i, v := range b {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf = append(buf, '-')
		}
		buf = append(buf, hexDigits[v>>4], hexDigits[v&0x0f])
	}
	return string(buf)
}

// newUUIDv7String - удобная обёртка.
func newUUIDv7String(timestampUnixMilli int64, r *rand.Rand) string {
	return formatUUID(newUUIDv7(timestampUnixMilli, r))
}
