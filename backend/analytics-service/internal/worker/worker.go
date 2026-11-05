package syncer

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/internal/services"
	"analytics-service/internal/storage"
)

type Service struct {
	logger          *slog.Logger
	cfg             config.SyncConfig
	userStorage     *storage.UserStorage
	paymentStorage  *storage.PaymentStorage
	subStorage      *storage.SubscriptionStorage
	planStorage     *storage.PlanStorage
	usersService    *services.UserService
	paymentsService *services.PaymentService
	subsService     *services.SubscriptionService
	metricsSvc      *services.MetricsService

	lastPoll    time.Time
	nextPoll    time.Time
	lastSyncErr error
	ready       atomic.Bool

	nextMu sync.Mutex
}

func New(
	logger *slog.Logger,
	cfg config.SyncConfig,
	userStorage *storage.UserStorage,
	paymentStorage *storage.PaymentStorage,
	subStorage *storage.SubscriptionStorage,
	planStorage *storage.PlanStorage,
	usersService *services.UserService,
	paymentsService *services.PaymentService,
	subsService *services.SubscriptionService,
	metricsSvc *services.MetricsService,
) *Service {
	return &Service{
		logger:          logger,
		cfg:             cfg,
		userStorage:     userStorage,
		paymentStorage:  paymentStorage,
		subStorage:      subStorage,
		planStorage:     planStorage,
		usersService:    usersService,
		paymentsService: paymentsService,
		subsService:     subsService,
		metricsSvc:      metricsSvc,
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
		case <-timer.C:
			timer.Reset(s.cfg.Interval)
			s.poll(ctx)
		}
	}
}

func (s *Service) NextPoll() time.Time {
	s.nextMu.Lock()
	defer s.nextMu.Unlock()
	return s.nextPoll
}

func (s *Service) setNextPoll(t time.Time) {
	s.nextMu.Lock()
	defer s.nextMu.Unlock()
	s.nextPoll = t
}

func (s *Service) Status() (lastPoll, nextPoll time.Time, ready bool, lastSyncErr string) {
	s.nextMu.Lock()
	lastPoll = s.lastPoll
	if s.lastSyncErr != nil {
		lastSyncErr = s.lastSyncErr.Error()
	}
	s.nextMu.Unlock()

	nextPoll = s.NextPoll()
	ready = s.ready.Load()

	return lastPoll, nextPoll, ready, lastSyncErr
}

func (s *Service) poll(ctx context.Context) {
	now := time.Now().UTC()
	s.setNextPoll(now.Add(s.cfg.Interval))

	var (
		wg       sync.WaitGroup
		errs     [4]error
		firstErr error
		users    []domain.User
		payments []domain.Payment
		subs     []domain.Subscription
		plans    []domain.Plan
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		users, errs[0] = s.usersService.FetchUsers(ctx, s.userStorage.UpdatedSince)
	}()
	go func() {
		defer wg.Done()
		payments, errs[1] = s.paymentsService.FetchPayments(ctx, s.paymentStorage.UpdatedSince)
	}()
	go func() {
		defer wg.Done()
		subs, errs[2] = s.subsService.FetchSubscriptions(ctx, s.subStorage.UpdatedSince)
	}()
	go func() {
		defer wg.Done()
		plans, errs[3] = s.subsService.FetchPlans(ctx)
	}()
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr != nil {
		s.nextMu.Lock()
		s.lastSyncErr = firstErr
		s.nextMu.Unlock()
		return
	}

	s.userStorage.Merge(users)
	s.paymentStorage.Merge(payments)
	s.subStorage.Merge(subs)
	s.planStorage.Replace(plans)

	s.paymentStorage.SetUpdatedSince(maxUpdatedOrDefault(s.paymentStorage.All(), func(p domain.Payment) time.Time { return p.UpdatedAt }))
	s.subStorage.SetUpdatedSince(maxUpdatedOrDefault(s.subStorage.All(), func(sub domain.Subscription) time.Time { return sub.UpdatedAt }))
	s.userStorage.SetUpdatedSince(maxUpdatedOrDefault(s.userStorage.All(), func(sub domain.User) time.Time { return sub.CreatedAt }))
	s.nextMu.Lock()
	s.lastPoll = time.Now().UTC()
	s.lastSyncErr = nil
	s.nextMu.Unlock()
	s.ready.Store(true)

	usersNow, paymentsNow, subsNow, plansNow := s.userStorage.All(), s.paymentStorage.All(), s.subStorage.All(), s.planStorage.All()
	s.logger.Info("sync: poll done",
		"users", len(usersNow),
		"payments", len(paymentsNow),
		"subscriptions", len(subsNow),
		"plans", len(plansNow),
	)

	if err := s.metricsSvc.Recalculate(ctx, usersNow, paymentsNow, subsNow, time.Now().UTC()); err != nil {
		s.logger.Error("metrics: recalculation failed", "error", err)
	}
}

func maxUpdatedOrDefault[T any](rows map[string]T, updated func(T) time.Time) *time.Time {
	var maxTime time.Time
	for _, row := range rows {
		t := updated(row)
		if t.After(maxTime) {
			maxTime = t
		}
	}
	if maxTime.IsZero() {
		return nil
	}
	return &maxTime
}
