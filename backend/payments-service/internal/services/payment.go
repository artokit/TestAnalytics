package services

import (
	"context"
	"log/slog"
	"payments-service/internal/errors"
	"time"

	"payments-service/internal/domain"
	"payments-service/internal/handlers/request"
	"payments-service/internal/handlers/response"
)

type IPaymentRepo interface {
	GetPayments(ctx context.Context, limit int, offset int, updatedSince *time.Time) ([]domain.Payment, int64, error)
}

type PaymentService struct {
	repo   IPaymentRepo
	logger *slog.Logger
}

func NewPaymentService(repo IPaymentRepo, logger *slog.Logger) *PaymentService {
	return &PaymentService{repo: repo, logger: logger}
}

func (s *PaymentService) GetPayments(ctx context.Context, filter *request.PaymentFilter) (response.PaymentsResponsePaged, error) {
	page, pageSize := normalizePaymentFilter(filter)

	offset := (page - 1) * pageSize

	payments, total, err := s.repo.GetPayments(ctx, pageSize, offset, filter.UpdatedSince)
	if err != nil {
		s.logger.Error("repo error", "method", "GetPayments", "page", page, "page_size", pageSize, "error", err)
		return response.PaymentsResponsePaged{}, errors.ErrGetPayments
	}

	result := make([]response.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		result = append(result, mapToPaymentResponse(p))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return response.PaymentsResponsePaged{
		Payments:   result,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}, nil
}

func normalizePaymentFilter(filter *request.PaymentFilter) (page int, pageSize int) {
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

func mapToPaymentResponse(p domain.Payment) response.PaymentResponse {
	return response.PaymentResponse{
		Id:             p.Id,
		UserId:         p.UserId,
		SubscriptionId: p.SubscriptionId,
		Amount:         p.Amount,
		Currency:       p.Currency,
		Status:         p.Status,
		PeriodStart:    p.PeriodStart,
		PeriodEnd:      p.PeriodEnd,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
