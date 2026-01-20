package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Gustik/gophermart/internal/model"
)

// GetOrderByNumber возвращает заказ по номеру
func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual
		FROM orders
		WHERE number = $1
	`

	var order model.Order
	err := s.db.QueryRowContext(ctx, query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("не удалось получить заказ: %w", err)
	}

	return &order, nil
}

// CreateOrder создает запись номера заказа
func (s *Storage) CreateOrder(ctx context.Context, userID int, number string) error {
	query := `
        INSERT INTO orders (user_id, number, status)
        VALUES ($1, $2, $3)
    `

	_, err := s.db.ExecContext(ctx, query, userID, number, model.OrderStatusNew)
	if err != nil {
		return fmt.Errorf("ошибка создания записи номера заказа: %w", err)
	}

	return nil
}
