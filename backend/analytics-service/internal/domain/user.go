package domain

import "time"

type User struct {
	Id         string    `json:"id"`
	TelegramId int64     `json:"telegram_id"`
	CreatedAt  time.Time `json:"created_at"`
	UtmLink    string    `json:"utm_link"`
}

type UsersPage struct {
	Users      []User `json:"users"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}
