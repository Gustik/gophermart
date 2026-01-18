package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Gustik/gophermart/internal/models"
)

// CreateUser создаёт нового пользователя в БД
func (s *Storage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, created_at
	`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return &user, nil
}

// GetUserByLogin получает пользователя по логину
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return &user, nil
}
