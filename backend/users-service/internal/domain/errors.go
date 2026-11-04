package domain

import "errors"

var (
	ErrInvalidParam = errors.New("неверно передан параметр")
	ErrGetUsers     = errors.New("ошибка при получение пользователей")
)
