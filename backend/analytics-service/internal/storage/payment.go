package storage

import (
	"sync"
	"time"

	"analytics-service/internal/domain"
)

type PaymentStorage struct {
	mu           sync.RWMutex
	Payments     map[string]domain.Payment
	UpdatedSince *time.Time
}

func NewPaymentStorage() *PaymentStorage {
	return &PaymentStorage{Payments: make(map[string]domain.Payment)}
}

func (s *PaymentStorage) Merge(payments []domain.Payment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range payments {
		s.Payments[p.Id] = p
	}
}

func (s *PaymentStorage) All() map[string]domain.Payment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Payments
}

func (s *PaymentStorage) SetUpdatedSince(t *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UpdatedSince = t
}
