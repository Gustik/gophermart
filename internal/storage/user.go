package storage

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/Gustik/gophermart/internal/model"
)

// CreateUser создаёт нового пользователя в БД
func (s *storage) CreateUser(ctx context.Context, login, passwordHash string) (*model.User, error) {
	query, args, err := s.psql.
		Insert("users").
		Columns("login", "password_hash").
		Values(login, passwordHash).
		Suffix("RETURNING id, login, created_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var user model.User
	err = s.db.GetContext(ctx, &user, query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return &user, nil
}

// GetUserByLogin получает пользователя по логину
func (s *storage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query, args, err := s.psql.
		Select("id", "login", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"login": login}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("не удалось построить запрос: %w", err)
	}

	var user model.User
	err = s.db.GetContext(ctx, &user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return &user, nil
}
