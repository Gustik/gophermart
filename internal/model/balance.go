package model

import "time"

// Balance представляет баланс пользователя
type Balance struct {
	UserID    int       `json:"-"`
	Current   float32   `json:"current"`   // текущий баланс баллов
	Withdrawn float32   `json:"withdrawn"` // сумма использованных баллов
	UpdatedAt time.Time `json:"-"`
}

// Withdrawal представляет операцию списания баллов
type Withdrawal struct {
	ID          int       `json:"-" db:"id"`
	UserID      int       `json:"-" db:"user_id"`
	OrderNumber string    `json:"order" db:"order_number"`
	Sum         float32   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"-" db:"processed_at"`
}
