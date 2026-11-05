package syncer

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

const (
	retryAttempts = 4
	backoffBase   = 250 * time.Millisecond
)

func requestWithRetry[T any](ctx context.Context, do func() (T, error)) (T, error) {
	var out T
	var lastErr error

	for attempt := range retryAttempts {
		if attempt > 0 {
			backoff := backoffBase << (attempt - 1)
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			select {
			case <-ctx.Done():
				return out, ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}

		out, lastErr = do()
		if lastErr == nil {
			return out, nil
		}
	}

	return out, fmt.Errorf("after %d attempts: %w", retryAttempts, lastErr)
}
