package storage

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/Gustik/gophermart/internal/model"
)

// GetOrderByNumber возвращает заказ по номеру
func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	query, args, err := s.psql.
		Select("id", "user_id", "number", "status", "accrual").
		From("orders").
		Where(sq.Eq{"number": number}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var order model.Order
	err = s.db.GetContext(ctx, &order, query, args...)
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
	query, args, err := s.psql.
		Insert("orders").
		Columns("user_id", "number", "status").
		Values(userID, number, model.OrderStatusNew).
		ToSql()

	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка создания записи номера заказа: %w", err)
	}

	return nil
}
