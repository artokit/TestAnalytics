package domain

import "time"

type User struct {
	Id         string    `db:"id"`
	TelegramId int64     `db:"telegram_id"`
	CreatedAt  time.Time `db:"created_at"`
	UtmLink    string    `db:"utm_link"`
}
