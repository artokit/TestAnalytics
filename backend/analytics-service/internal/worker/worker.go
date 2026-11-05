package syncer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/internal/services"
)

const pageSize = 100

type Service struct {
	logger          *slog.Logger
	cfg             config.SyncConfig
	usersService    *services.UserService
	paymentsService *services.PaymentService
	subsService     *services.SubscriptionService
	planService     *services.PlanService
	metricsSvc      *services.MetricsService

	users      map[string]domain.User
	payments   map[string]domain.Payment
	subs       map[string]domain.Subscription
	plans      map[int]domain.Plan
	watermark  *time.Time
	subsWatermark *time.Time
	ready      atomic.Bool

	lastPoll    time.Time
	nextPoll    time.Time
	lastSyncErr string

	readCh chan readRequest
}

type readRequest struct {
	fn   func(s *Service)
	done chan struct{}
}

func New(logger *slog.Logger, cfg config.SyncConfig, usersService *services.UserService, paymentsService *services.PaymentService, subsService *services.SubscriptionService, planService *services.PlanService, metricsSvc *services.MetricsService) *Service {
	return &Service{
		logger:          logger,
		cfg:             cfg,
		usersService:    usersService,
		paymentsService: paymentsService,
		subsService:     subsService,
		planService:     planService,
		metricsSvc:      metricsSvc,
		users:           make(map[string]domain.User),
		payments:        make(map[string]domain.Payment),
		subs:            make(map[string]domain.Subscription),
		plans:           make(map[int]domain.Plan),
		readCh:          make(chan readRequest),
	}
}

func (s *Service) Run(ctx context.Context) {
	s.poll(ctx)

	timer := time.NewTimer(s.cfg.Interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-s.readCh:
			req.exec(s)
			close(req.done)
		case <-timer.C:
			timer.Reset(s.cfg.Interval)
			s.poll(ctx)
		}
	}
}

func (s *Service) Read(fn func(users map[string]domain.User, payments map[string]domain.Payment, subs map[string]domain.Subscription, plans map[int]domain.Plan)) {
	done := make(chan struct{})
	s.readCh <- readRequest{fn: func(s *Service) { fn(s.users, s.payments, s.subs, s.plans) }, done: done}
	<-done
}

func (s *Service) Status() (lastPoll, nextPoll time.Time, lastErr string) {
	done := make(chan struct{})
	s.readCh <- readRequest{fn: func(s *Service) { lastPoll, nextPoll, lastErr = s.lastPoll, s.nextPoll, s.lastSyncErr }, done: done}
	<-done
	return
}

func (r readRequest) exec(s *Service) {
	r.fn(s)
}

func (s *Service) poll(ctx context.Context) {
	now := time.Now().UTC()
	s.nextPoll = now.Add(s.cfg.Interval)

	var (
		wg   sync.WaitGroup
		errs [4]error
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		errs[0] = s.fetchUsers(ctx)
	}()
	go func() {
		defer wg.Done()
		errs[1] = s.fetchPayments(ctx)
	}()
	go func() {
		defer wg.Done()
		errs[2] = s.fetchSubscriptions(ctx)
	}()
	go func() {
		defer wg.Done()
		errs[3] = s.fetchPlans(ctx)
	}()
	wg.Wait()

	var firstErr error
	for _, err := range errs {
		if err != nil {
			firstErr = err
			s.logger.Error("sync: source poll failed, cache is kept", "error", err)
		}
	}
	if firstErr != nil {
		s.lastSyncErr = firstErr.Error()
		return
	}

	s.watermark = maxUpdatedOrDefault(s.payments, func(p domain.Payment) time.Time { return p.UpdatedAt })
	s.subsWatermark = maxUpdatedOrDefault(s.subs, func(sub domain.Subscription) time.Time { return sub.UpdatedAt })
	s.ready.Store(true)
	s.lastPoll = time.Now().UTC()
	s.lastSyncErr = ""

	s.logger.Info("sync: poll done",
		"users", len(s.users),
		"payments", len(s.payments),
		"subscriptions", len(s.subs),
		"plans", len(s.plans),
	)

	if err := s.metricsSvc.Recalculate(ctx, s.users, s.payments, s.subs, time.Now().UTC()); err != nil {
		s.logger.Error("metrics: recalculation failed", "error", err)
	}
}

func (s *Service) fetchSubscriptions(ctx context.Context) error {
	urlFor := func(page int) string {
		u := fmt.Sprintf("%s?page=%d&page_size=%d", s.cfg.SubscriptionsURL, page, pageSize)
		if s.subsWatermark != nil {
			u += "&updated_since=" + s.subsWatermark.Format(time.RFC3339)
		}
		return u
	}

	get := func(ctx context.Context, url string) (domain.SubscriptionsPage, error) {
		resp, err := s.subsService.GetSubscriptionsPage(ctx, url)
		if err != nil {
			return domain.SubscriptionsPage{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return domain.SubscriptionsPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}

	selectRows := func(p domain.SubscriptionsPage) ([]domain.Subscription, int) {
		return p.Subscriptions, p.TotalPages
	}

	apply := func(sub domain.Subscription) {
		s.subs[sub.Id] = sub
	}

	return s.fetchPages[domain.SubscriptionsPage](ctx, urlFor, get, selectRows, apply)
}

func (s *Service) fetchPlans(ctx context.Context) error {
	get := func(ctx context.Context, url string) (domain.PlansResponse, error) {
		resp, err := s.planService.GetPlans(ctx, url)
		if err != nil {
			return domain.PlansResponse{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return domain.PlansResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}

	p, err := requestWithRetry[domain.PlansResponse](ctx, func() (domain.PlansResponse, error) {
		return get(ctx, s.cfg.PlansURL)
	})
	if err != nil {
		return err
	}

	for _, plan := range p.Items {
		s.plans[plan.Id] = plan
	}

	return nil
}

func (s *Service) fetchUsers(ctx context.Context) error {
	urlFor := func(page int) string {
		return fmt.Sprintf("%s?page=%d&page_size=%d", s.cfg.UsersURL, page, pageSize)
	}

	get := func(ctx context.Context, url string) (domain.UsersPage, error) {
		resp, err := s.usersService.GetUsersPage(ctx, url)
		if err != nil {
			return domain.UsersPage{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return domain.UsersPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}

	selectRows := func(p domain.UsersPage) ([]domain.User, int) {
		return p.Users, p.TotalPages
	}

	apply := func(u domain.User) {
		s.users[u.Id] = u
	}

	return s.fetchPages[domain.UsersPage](ctx, urlFor, get, selectRows, apply)
}

func (s *Service) fetchPayments(ctx context.Context) error {
	urlFor := func(page int) string {
		u := fmt.Sprintf("%s?page=%d&page_size=%d", s.cfg.PaymentsURL, page, pageSize)
		if s.watermark != nil {
			u += "&updated_since=" + s.watermark.Format(time.RFC3339)
		}
		return u
	}

	get := func(ctx context.Context, url string) (domain.PaymentsPage, error) {
		resp, err := s.paymentsService.GetPaymentsPage(ctx, url)
		if err != nil {
			return domain.PaymentsPage{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return domain.PaymentsPage{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}

	selectRows := func(p domain.PaymentsPage) ([]domain.Payment, int) {
		return p.Payments, p.TotalPages
	}

	apply := func(p domain.Payment) {
		s.payments[p.Id] = p
	}

	return s.fetchPages[domain.PaymentsPage](ctx, urlFor, get, selectRows, apply)
}

func (s *Service) fetchPages[P, Row any](
	ctx context.Context,
	urlFor func(page int) string,
	get func(ctx context.Context, url string) (P, error),
	selectRows func(P) ([]Row, int),
	apply func(Row),
) error {
	for page := 1; ; page++ {
		var zero P
		url := urlFor(page)
		p, err := requestWithRetry[P](ctx, func() (P, error) {
			resp, err := get(ctx, url)
			if err != nil {
				return zero, err
			}
			return resp, nil
		})
		if err != nil {
			return err
		}

		rows, total := selectRows(p)
		for _, r := range rows {
			apply(r)
		}
		if page >= total {
			break
		}
	}

	return nil
}

func (s *Service) Ready() bool {
	return s.ready.Load()
}

func maxUpdatedOrDefault[T any](rows map[string]T, updated func(T) time.Time) *time.Time {
	var max time.Time
	for _, row := range rows {
		t := updated(row)
		if t.After(max) {
			max = t
		}
	}
	if max.IsZero() {
		return nil
	}
	return &max
}
