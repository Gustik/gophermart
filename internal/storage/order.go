package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/Gustik/gophermart/internal/model"
)

var (
	ErrOrderNotFound      = errors.New("заказ не найден")
	ErrOrderAlreadyExists = errors.New("заказ уже существует")
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
			return nil, ErrOrderNotFound
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
		// Проверяем на unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == pgDuplicateErrorCode { // unique_violation
				return ErrOrderAlreadyExists
			}
		}
		return fmt.Errorf("ошибка создания записи номера заказа: %w", err)
	}

	return nil
}

// GetPendingOrders возвращает заказы, ожидающие обработки (NEW, PROCESSING)
func (s *Storage) GetPendingOrders(ctx context.Context) ([]model.Order, error) {
	query, args, err := s.psql.
		Select("id", "user_id", "number", "status", "accrual").
		From("orders").
		Where(sq.Eq{"status": []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var orders []model.Order
	err = s.db.SelectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказы: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус и сумму начисления для заказа
func (s *Storage) UpdateOrderStatus(ctx context.Context, orderNumber string, status model.OrderStatus, accrual *float32) error {
	updateBuilder := s.psql.
		Update("orders").
		Set("status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"number": orderNumber})

	if accrual != nil {
		updateBuilder = updateBuilder.Set("accrual", *accrual)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус заказа: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// GetOrdersByUser возвращает заказыва пользователя
func (s *Storage) GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error) {
	query, args, err := s.psql.
		Select("id", "user_id", "number", "status", "accrual", "uploaded_at").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("uploaded_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var orders []model.Order
	err = s.db.SelectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказы: %w", err)
	}

	return orders, nil
}
