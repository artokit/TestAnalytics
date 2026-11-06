package domain

import "errors"

var (
	ErrGetCurrencyPrices = errors.New("ошибка при получении цен валют")
	ErrGetUserLtv        = errors.New("ошибка при получении LTV пользователя")
	ErrUserNotFound      = errors.New("пользователь с таким telegram_id не найден")
)
