package storage

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/Gustik/gophermart/internal/model"
)

// InitBalance инициализирует баланс для нового пользователя
func (s *Storage) InitBalance(ctx context.Context, userID int) error {
	query, args, err := s.psql.
		Insert("balance").
		Columns("user_id", "current", "withdrawn").
		Values(userID, 0, 0).
		ToSql()

	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось создать баланс: %w", err)
	}

	return nil
}

// AddBalance добавляет баллы к балансу пользователя
func (s *Storage) AddBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount *float32) error {
	query, args, err := s.psql.
		Update("balance").
		Set("current", sq.Expr("current + ?", amount)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	_, err = s.getExecutor(tx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс: %w", err)
	}

	return nil
}

// GetBalance возвращает текущий баланс пользователя
func (s *Storage) GetBalance(ctx context.Context, tx *sqlx.Tx, userID int) (*model.Balance, error) {
	builder := s.psql.
		Select("user_id", "current", "withdrawn", "updated_at").
		From("balance").
		Where(sq.Eq{"user_id": userID})

	if tx != nil {
		builder = builder.Suffix("FOR UPDATE")
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var balance model.Balance
	err = s.getExecutor(tx).QueryRowxContext(ctx, query, args...).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
		&balance.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить баланс: %w", err)
	}

	return &balance, nil
}

// WithdrawBalance списывает средства с баланса пользователя
func (s *Storage) WithdrawBalance(ctx context.Context, tx *sqlx.Tx, userID int, sum float32) error {
	query, args, err := s.psql.
		Update("balance").
		Set("current", sq.Expr("current - ?", sum)).
		Set("withdrawn", sq.Expr("withdrawn + ?", sum)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	_, err = s.getExecutor(tx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось списать средства: %w", err)
	}

	return nil
}

// CreateWithdrawal создаёт запись о списании баллов
func (s *Storage) CreateWithdrawal(ctx context.Context, tx *sqlx.Tx, userID int, orderNumber string, sum float32) error {
	query, args, err := s.psql.
		Insert("withdrawals").
		Columns("user_id", "order_number", "sum").
		Values(userID, orderNumber, sum).
		ToSql()

	if err != nil {
		return fmt.Errorf("не удалось построить запрос: %w", err)
	}

	_, err = s.getExecutor(tx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("не удалось создать запись о списании: %w", err)
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя, отсортированную от новых к старым
func (s *Storage) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	query, args, err := s.psql.
		Select("id", "user_id", "order_number", "sum", "processed_at").
		From("withdrawals").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("processed_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var withdrawals []model.Withdrawal
	err = s.db.SelectContext(ctx, &withdrawals, query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список списаний: %w", err)
	}

	return withdrawals, nil
}
