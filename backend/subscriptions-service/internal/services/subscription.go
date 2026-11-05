package services

import (
	"context"
	"log/slog"
	"time"

	"subscriptions-service/internal/domain"
	subscription_errors "subscriptions-service/internal/errors/subscription"
	"subscriptions-service/internal/handlers/request"
	"subscriptions-service/internal/handlers/response"
)

type ISubscriptionRepo interface {
	GetSubscriptions(ctx context.Context, limit int, offset int, updatedSince *time.Time) ([]domain.Subscription, int64, error)
}

type SubscriptionService struct {
	repo   ISubscriptionRepo
	logger *slog.Logger
}

func NewSubscriptionService(repo ISubscriptionRepo, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func (s *SubscriptionService) GetSubscriptions(ctx context.Context, filter *request.SubscriptionFilter) (response.SubscriptionsResponsePaged, error) {
	page, pageSize := normalizeSubscriptionFilter(filter)

	offset := (page - 1) * pageSize

	subs, total, err := s.repo.GetSubscriptions(ctx, pageSize, offset, filter.UpdatedSince)
	if err != nil {
		s.logger.Error("repo error", "method", "GetSubscriptions", "page", page, "page_size", pageSize, "error", err)
		return response.SubscriptionsResponsePaged{}, subscription_errors.ErrGetSubscriptions
	}

	result := make([]response.SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		result = append(result, mapToSubscriptionResponse(sub))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return response.SubscriptionsResponsePaged{
		Subscriptions: result,
		Total:         int(total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int(totalPages),
	}, nil
}

func normalizeSubscriptionFilter(filter *request.SubscriptionFilter) (page int, pageSize int) {
	if filter == nil {
		return request.DefaultPage, request.DefaultPageSize
	}

	page = max(filter.Page, request.MinPage)

	pageSize = filter.PageSize
	if pageSize < request.MinPageSize {
		pageSize = request.DefaultPageSize
	}

	return page, pageSize
}

func mapToSubscriptionResponse(sub domain.Subscription) response.SubscriptionResponse {
	return response.SubscriptionResponse{
		Id:          sub.Id,
		UserId:      sub.UserId,
		PlanId:      sub.PlanId,
		PeriodStart: sub.PeriodStart,
		PeriodEnd:   sub.PeriodEnd,
		Status:      sub.Status,
		CreatedAt:   sub.CreatedAt,
		UpdatedAt:   sub.UpdatedAt,
	}
}
