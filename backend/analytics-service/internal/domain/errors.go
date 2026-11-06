package domain

import "errors"

var (
	ErrInvalidParam      = errors.New("неверно передан параметр")
	ErrGetCurrencyPrices = errors.New("ошибка при получении цен валют")
)
