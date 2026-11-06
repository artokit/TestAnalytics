package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"analytics-service/internal/domain"
)

func seedUsers() []domain.User {
	users := make([]domain.User, 0, 32)
	for i := 0; i < 32; i++ {
		users = append(users, domain.User{Id: fmt.Sprintf("u%02d", i), TelegramId: int64(100500 + i)})
	}
	return users
}

func seedPayments() []domain.Payment {
	payments := make([]domain.Payment, 0, 32)
	for i := 0; i < 32; i++ {
		payments = append(payments, domain.Payment{
			Id:     fmt.Sprintf("p%02d", i),
			UserId: fmt.Sprintf("u%02d", i%8),
			Amount: int64(1000 + i),
			Status: "succeeded",
		})
	}
	return payments
}

func seedSubs() []domain.Subscription {
	subs := make([]domain.Subscription, 0, 32)
	for i := 0; i < 32; i++ {
		base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * 30 * 24 * time.Hour)
		subs = append(subs, domain.Subscription{
			Id:          fmt.Sprintf("s%02d", i),
			UserId:      fmt.Sprintf("u%02d", i%8),
			PlanId:      1,
			PeriodStart: base,
			PeriodEnd:   base.Add(30 * 24 * time.Hour),
			Status:      "active",
		})
	}
	return subs
}

func TestUserStorageConcurrentMergeRead(t *testing.T) {
	s := NewUserStorage()
	rows := seedUsers()

	s.Merge(rows)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := 0; r < 40; r++ {
			for id, u := range s.All() {
				if u.TelegramId < 0 || id == "" {
					t.Errorf("corrupt user id=%q telegram=%d", id, u.TelegramId)
				}
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()

	for i, u := range rows {
		u.CreatedAt = u.CreatedAt.Add(time.Duration(i) * time.Second)
		s.Merge([]domain.User{u})
		time.Sleep(500 * time.Microsecond)
	}

	wg.Wait()

	if len(s.All()) != len(rows) {
		t.Fatalf("final map size = %d, want %d", len(s.All()), len(rows))
	}
}

func TestPaymentStorageConcurrentMergeRead(t *testing.T) {
	s := NewPaymentStorage()
	rows := seedPayments()

	s.Merge(rows)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := 0; r < 40; r++ {
			for id, p := range s.All() {
				if p.UserId == "" || id == "" {
					t.Errorf("corrupt payment id=%q", id)
				}
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()

	for i, p := range rows {
		p.Amount += int64(i)
		s.Merge([]domain.Payment{p})
		time.Sleep(500 * time.Microsecond)
	}

	wg.Wait()

	if len(s.All()) != len(rows) {
		t.Fatalf("final map size = %d, want %d", len(s.All()), len(rows))
	}
}

func TestSubscriptionStorageConcurrentMergeRead(t *testing.T) {
	s := NewSubscriptionStorage()
	rows := seedSubs()

	s.Merge(rows)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := 0; r < 40; r++ {
			for id, sub := range s.All() {
				if sub.UserId == "" || id == "" {
					t.Errorf("corrupt subscription id=%q", id)
				}
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()

	for i, sub := range rows {
		sub.PlanId = int16(1 + i%4)
		s.Merge([]domain.Subscription{sub})
		time.Sleep(500 * time.Microsecond)
	}

	wg.Wait()

	if len(s.All()) != len(rows) {
		t.Fatalf("final map size = %d, want %d", len(s.All()), len(rows))
	}
}

func TestPlanStorageConcurrentReplaceRead(t *testing.T) {
	s := NewPlanStorage()
	rows := []domain.Plan{
		{Id: 1, PeriodMonths: 1, Price: 50000},
		{Id: 2, PeriodMonths: 3, Price: 135000},
		{Id: 3, PeriodMonths: 6, Price: 240000},
		{Id: 4, PeriodMonths: 12, Price: 420000},
	}

	s.Replace(rows)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := 0; r < 40; r++ {
			for id, plan := range s.All() {
				if id != int(plan.Id) {
					t.Errorf("plan id mismatch id=%d plan.Id=%d", id, plan.Id)
				}
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()

	for i := 0; i < 40; i++ {
		s.Replace(rows)
		time.Sleep(500 * time.Microsecond)
	}

	wg.Wait()

	if len(s.All()) != len(rows) {
		t.Fatalf("final map size = %d, want %d", len(s.All()), len(rows))
	}
}
