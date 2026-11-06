package middleware

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mock_http_requests_total",
		Help: "Total number of HTTP requests handled by mock services",
	}, []string{"service"})

	httpErrorResponsesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mock_http_error_responses_total",
		Help: "Error responses (429 and 500) returned by mock services, by status code",
	}, []string{"service", "code"})
)

type Metrics struct {
	service string
}

func NewMetrics(service string) *Metrics {
	return &Metrics{service: service}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		httpRequestsTotal.WithLabelValues(m.service).Inc()

		switch recorder.status {
		case http.StatusTooManyRequests, http.StatusInternalServerError:
			httpErrorResponsesTotal.WithLabelValues(m.service, strconv.Itoa(recorder.status)).Inc()
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
