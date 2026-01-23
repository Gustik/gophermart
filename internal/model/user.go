package model

import "time"

// User представляет пользователя системы
type User struct {
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	PasswordHash string    `json:"-" db:"password_hash"` // не включаем в JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
