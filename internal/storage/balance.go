package storage

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

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
