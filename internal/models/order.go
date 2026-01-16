package models

import "time"

// OrderStatus представляет статус заказа
type OrderStatus string

const (
	// OrderStatusNew - заказ загружен в систему
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessing - вознаграждение за заказ рассчитывается
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusInvalid - система расчёта отказала в расчёте
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusProcessed - данные по заказу проверены
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// Order представляет заказ пользователя
type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float32    `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
