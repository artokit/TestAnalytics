package services

import (
	"context"

	"analytics-service/config"
	"analytics-service/internal/domain"
	"analytics-service/pkg"
)

type UserService struct {
	hc     *httpclient.HttpClient
	config *config.Config
}

func NewUserService(hc *httpclient.HttpClient, cfg *config.Config) *UserService {
	return &UserService{
		hc:     hc,
		config: cfg,
	}
}

func (s *UserService) GetUsersPage(ctx context.Context, url string) (httpclient.Response[domain.UsersPage], error) {
	return s.hc.Get[domain.UsersPage](ctx, url)
}
