package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/storage"
)

var (
	ErrUserAlreadyExists  = errors.New("пользователь уже существует")
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
)

// AuthService определяет методы для сервиса аутентификации
type AuthService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// authService предоставляет бизнес-логику для аутентификации
type authService struct {
	jwtSecret string
	storage   *storage.Storage
}

// NewAuthService создаёт новый AuthService
func NewAuthService(jwtSecret string, storage *storage.Storage) AuthService {
	return &authService{
		jwtSecret: jwtSecret,
		storage:   storage,
	}
}

// Register регистрирует нового пользователя
func (s *authService) Register(ctx context.Context, login, password string) (string, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("ошибка проверки пользователя: %w", err)
	}
	if existingUser != nil {
		return "", ErrUserAlreadyExists
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// Создаём пользователя
	user, err := s.storage.CreateUser(ctx, login, string(passwordHash))
	if err != nil {
		return "", fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	// Создаём баланс для пользователя
	if err := s.storage.CreateBalance(ctx, user.ID); err != nil {
		return "", fmt.Errorf("ошибка создания баланса: %w", err)
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return token, nil
}

// Login аутентифицирует пользователя
func (s *authService) Login(ctx context.Context, login, password string) (string, error) {
	// Получаем пользователя из БД
	user, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(s.jwtSecret, user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return token, nil
}
