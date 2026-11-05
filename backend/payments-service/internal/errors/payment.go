package errors

import "errors"

var (
	ErrInvalidQueryParameter = errors.New("неверно передан параметр")
	ErrGetPayments           = errors.New("ошибка при получении платежей")
)
