package services

import (
	"context"
	"log/slog"

	"users-service/internal/domain"
	"users-service/internal/handlers/request"
	"users-service/internal/handlers/response"
)

type IUserRepo interface {
	GetUsers(ctx context.Context, limit int, offset int) ([]domain.User, int64, error)
}

type UserService struct {
	repo   IUserRepo
	logger *slog.Logger
}

func New(repo IUserRepo, logger *slog.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) GetUsers(ctx context.Context, filter *request.UserFilter) (response.UsersResponsePaged, error) {
	page, pageSize := normalizeFilter(filter)

	offset := (page - 1) * pageSize

	users, total, err := s.repo.GetUsers(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error("repo error", "method", "GetUsers", "page", page, "page_size", pageSize, "error", err)
		return response.UsersResponsePaged{}, domain.ErrGetUsers
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	if len(users) == 0 {
		return response.UsersResponsePaged{
			Users:      []response.UserResponse{},
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: int(totalPages),
		}, nil
	}

	result := make([]response.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, mapToResponse(u))
	}

	return response.UsersResponsePaged{
		Users:      result,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}, nil
}

func normalizeFilter(filter *request.UserFilter) (page int, pageSize int) {
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

func mapToResponse(user domain.User) response.UserResponse {
	return response.UserResponse{
		Id:         user.Id,
		TelegramId: user.TelegramId,
		CreatedAt:  user.CreatedAt,
		UtmLink:    user.UtmLink,
	}
}
