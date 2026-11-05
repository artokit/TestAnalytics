package handlers

import (
	"context"
	"net/http"
	"shared/httputil"
	"users-service/internal/handlers/request"
	"users-service/internal/handlers/response"
)

type IUserService interface {
	GetUsers(ctx context.Context, filter *request.UserFilter) (response.UsersResponsePaged, error)
}

func GetUsers(service IUserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := request.GetUserFilter(r)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		res, err := service.GetUsers(r.Context(), filter)
		if err != nil {
			httputil.WriteError(w, err, http.StatusBadRequest)
			return
		}

		httputil.WriteJSON(w, res, http.StatusOK)
	}
}
