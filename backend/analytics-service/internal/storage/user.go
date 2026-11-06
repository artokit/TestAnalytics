package storage

import (
	"sync"
	"time"

	"analytics-service/internal/domain"
)

type UserStorage struct {
	mu           sync.RWMutex
	Users        map[string]domain.User
	UpdatedSince *time.Time
}

func NewUserStorage() *UserStorage {
	return &UserStorage{Users: make(map[string]domain.User)}
}

func (s *UserStorage) Merge(users []domain.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range users {
		s.Users[u.Id] = u
	}
}

func (s *UserStorage) All() map[string]domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]domain.User, len(s.Users))
	for id, u := range s.Users {
		out[id] = u
	}
	return out
}

func (s *UserStorage) SetUpdatedSince(t *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UpdatedSince = t
}
