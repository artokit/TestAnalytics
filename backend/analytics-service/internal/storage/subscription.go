package storage

import (
	"sync"
	"time"

	"analytics-service/internal/domain"
)

type SubscriptionStorage struct {
	mu            sync.RWMutex
	Subscriptions map[string]domain.Subscription
	UpdatedSince  *time.Time
}

func NewSubscriptionStorage() *SubscriptionStorage {
	return &SubscriptionStorage{Subscriptions: make(map[string]domain.Subscription)}
}

func (s *SubscriptionStorage) Merge(subs []domain.Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sub := range subs {
		s.Subscriptions[sub.Id] = sub
	}
}

func (s *SubscriptionStorage) All() map[string]domain.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]domain.Subscription, len(s.Subscriptions))
	for id, sub := range s.Subscriptions {
		out[id] = sub
	}
	return out
}

func (s *SubscriptionStorage) SetUpdatedSince(t *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UpdatedSince = t
}
