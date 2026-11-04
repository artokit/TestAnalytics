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

// GetUsers godoc
// @Summary      Список пользователей
// @Description  Возвращает пользователей постранично (пагинация через query-параметры)
// @Tags         users
// @Produce      json
// @Param        page       query integer false "Номер страницы (минимум 1)"     minimum(1) default(1)
// @Param        page_size  query integer false "Размер страницы (1-100)"  minimum(1) maximum(100) default(50)
// @Success      200 {object}  response.UsersResponsePaged
// @Failure      400 {object} httputil.ErrorResponse
// @Router       /users [get]
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
