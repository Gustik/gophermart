package model

import "time"

// Balance представляет баланс пользователя
type Balance struct {
	UserID    int       `json:"user_id"`
	Current   float32   `json:"current"`   // текущий баланс баллов
	Withdrawn float32   `json:"withdrawn"` // сумма использованных баллов
	UpdatedAt time.Time `json:"updated_at"`
}

// Withdrawal представляет операцию списания баллов
type Withdrawal struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
