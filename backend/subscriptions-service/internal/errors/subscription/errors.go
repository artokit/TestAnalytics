package subscription

import "errors"

var (
	ErrInvalidPage       = errors.New("неверно передан параметр page")
	ErrInvalidPageSize   = errors.New("неверно передан параметр page_size")
	ErrInvalidUpdatedSince = errors.New("неверно передан параметр updated_since, expected RFC3339")
	ErrGetSubscriptions  = errors.New("ошибка получения подписок")
)
