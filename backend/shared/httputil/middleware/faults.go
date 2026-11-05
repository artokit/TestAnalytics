package middleware

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"shared/httputil"
	"time"
)

type FaultsConfig struct {
	ErrorRate500 float64
	ErrorRate429 float64
	RetryAfter   time.Duration
}

const (
	minDelay = 200 * time.Millisecond
	maxDelay = 500 * time.Millisecond
)

type Faults struct {
	config FaultsConfig
}

func NewFaults(errorRate500, errorRate429 float64, retryAfter time.Duration) *Faults {
	return &Faults{
		config: FaultsConfig{
			ErrorRate500: errorRate500,
			ErrorRate429: errorRate429,
			RetryAfter:   retryAfter,
		},
	}
}

func (f *Faults) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(randomDelay())

		switch status := randomFault(f.config); status {
		case 0:
			next.ServeHTTP(w, r)
		case http.StatusTooManyRequests:
			w.Header().Set("Retry-After", strconv.Itoa(int(f.config.RetryAfter.Seconds())))
			httputil.WriteError(w, errors.New("too many requests"), status)
		default:
			httputil.WriteError(w, errors.New("simulated internal server error"), status)
		}
	})
}

func randomFault(cfg FaultsConfig) int {
	roll := rand.Float64()
	switch {
	case roll < cfg.ErrorRate429:
		return http.StatusTooManyRequests
	case roll < cfg.ErrorRate429+cfg.ErrorRate500:
		return http.StatusInternalServerError
	default:
		return 0
	}
}

func randomDelay() time.Duration {
	return minDelay + time.Duration(rand.Int64N(int64(maxDelay-minDelay)))
}

func HealthSafe(faults func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}
			faults(next).ServeHTTP(w, r)
		})
	}
}
