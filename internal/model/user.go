package model

import "time"

// User представляет пользователя системы
type User struct {
	ID           int       `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"` // не включаем в JSON
	CreatedAt    time.Time `json:"created_at"`
}
