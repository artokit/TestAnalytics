package request

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetTelegramId(r *http.Request) (int64, error) {
	raw, ok := mux.Vars(r)["telegram_id"]
	if !ok {
		return 0, errors.New("не передан параметр telegram_id")
	}

	telegramId, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New("неверно передан параметр telegram_id, expected integer")
	}

	return telegramId, nil
}
