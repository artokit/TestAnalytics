package storage

import (
	"sync"

	"analytics-service/internal/domain"
)

type PlanStorage struct {
	mu    sync.RWMutex
	Plans map[int]domain.Plan
}

func NewPlanStorage() *PlanStorage {
	return &PlanStorage{Plans: make(map[int]domain.Plan)}
}

func (s *PlanStorage) Replace(plans []domain.Plan) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := make(map[int]domain.Plan, len(plans))
	for _, p := range plans {
		next[p.Id] = p
	}
	s.Plans = next
}

func (s *PlanStorage) All() map[int]domain.Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[int]domain.Plan, len(s.Plans))
	for id, p := range s.Plans {
		out[id] = p
	}
	return out
}
