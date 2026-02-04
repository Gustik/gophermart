package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/model"
)

func TestRegister_Success(t *testing.T) {
	login := "testuser"
	password := "testpassword"
	userID := 1

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			// Пользователь не существует
			return nil, nil
		},
		createUserFunc: func(ctx context.Context, login, passwordHash string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Login:        login,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now(),
			}, nil
		},
		initBalanceFunc: func(ctx context.Context, userID int) error {
			return nil
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Register(context.Background(), login, password)
	if err != nil {
		t.Errorf("Register() ошибка = %v, ожидалось nil", err)
		return
	}

	if token == "" {
		t.Error("Register() вернул пустой токен")
	}

	// Проверяем, что токен валидный
	claims, err := auth.ValidateToken("test-secret", token)
	if err != nil {
		t.Errorf("ValidateToken() ошибка = %v, токен должен быть валидным", err)
		return
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() userID = %v, ожидалось %v", claims.UserID, userID)
	}
	if claims.Login != login {
		t.Errorf("ValidateToken() login = %v, ожидалось %v", claims.Login, login)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	login := "existinguser"
	password := "testpassword"

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			// Пользователь уже существует
			return &model.User{
				ID:           1,
				Login:        login,
				PasswordHash: "somehash",
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Register(context.Background(), login, password)
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("Register() ошибка = %v, ожидалось ErrUserAlreadyExists", err)
	}
	if token != "" {
		t.Error("Register() токен должен быть пустым при ошибке")
	}
}

func TestRegister_GetUserByLoginError(t *testing.T) {
	login := "testuser"
	password := "testpassword"

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Register(context.Background(), login, password)
	if err == nil {
		t.Error("Register() ожидалась ошибка при ошибке GetUserByLogin")
	}
	if token != "" {
		t.Error("Register() токен должен быть пустым при ошибке")
	}
}

func TestRegister_CreateUserError(t *testing.T) {
	login := "testuser"
	password := "testpassword"

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, nil
		},
		createUserFunc: func(ctx context.Context, login, passwordHash string) (*model.User, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Register(context.Background(), login, password)
	if err == nil {
		t.Error("Register() ожидалась ошибка при ошибке CreateUser")
	}
	if token != "" {
		t.Error("Register() токен должен быть пустым при ошибке")
	}
}

func TestRegister_InitBalanceError(t *testing.T) {
	login := "testuser"
	password := "testpassword"
	userID := 1

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, nil
		},
		createUserFunc: func(ctx context.Context, login, passwordHash string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Login:        login,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now(),
			}, nil
		},
		initBalanceFunc: func(ctx context.Context, userID int) error {
			return errors.New("database error")
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Register(context.Background(), login, password)
	if err == nil {
		t.Error("Register() ожидалась ошибка при ошибке InitBalance")
	}
	if token != "" {
		t.Error("Register() токен должен быть пустым при ошибке")
	}
}

func TestLogin_Success(t *testing.T) {
	login := "testuser"
	password := "testpassword"
	userID := 1

	// Хешируем пароль для хранения в БД
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("не удалось захешировать пароль: %v", err)
	}

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Login:        login,
				PasswordHash: string(passwordHash),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Login(context.Background(), login, password)
	if err != nil {
		t.Errorf("Login() ошибка = %v, ожидалось nil", err)
		return
	}

	if token == "" {
		t.Error("Login() вернул пустой токен")
	}

	// Проверяем, что токен валидный
	claims, err := auth.ValidateToken("test-secret", token)
	if err != nil {
		t.Errorf("ValidateToken() ошибка = %v, токен должен быть валидным", err)
		return
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() userID = %v, ожидалось %v", claims.UserID, userID)
	}
	if claims.Login != login {
		t.Errorf("ValidateToken() login = %v, ожидалось %v", claims.Login, login)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	login := "nonexistentuser"
	password := "testpassword"

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			// Пользователь не найден
			return nil, nil
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Login(context.Background(), login, password)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() ошибка = %v, ожидалось ErrInvalidCredentials", err)
	}
	if token != "" {
		t.Error("Login() токен должен быть пустым при ошибке")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	login := "testuser"
	correctPassword := "correctpassword"
	wrongPassword := "wrongpassword"
	userID := 1

	// Хешируем правильный пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("не удалось захешировать пароль: %v", err)
	}

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Login:        login,
				PasswordHash: string(passwordHash),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Login(context.Background(), login, wrongPassword)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() ошибка = %v, ожидалось ErrInvalidCredentials", err)
	}
	if token != "" {
		t.Error("Login() токен должен быть пустым при ошибке")
	}
}

func TestLogin_GetUserByLoginError(t *testing.T) {
	login := "testuser"
	password := "testpassword"

	mock := &mockStorage{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewAuthService("test-secret", mock)

	token, err := service.Login(context.Background(), login, password)
	if err == nil {
		t.Error("Login() ожидалась ошибка при ошибке GetUserByLogin")
	}
	if token != "" {
		t.Error("Login() токен должен быть пустым при ошибке")
	}
}
