package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/service"
)

// MockAuthService для тестирования
type MockAuthService struct {
	RegisterFunc func(ctx context.Context, login, password string) (string, error)
	LoginFunc    func(ctx context.Context, login, password string) (string, error)
}

func (m *MockAuthService) Register(ctx context.Context, login, password string) (string, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, login, password)
	}
	return "", nil
}

func (m *MockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, login, password)
	}
	return "", nil
}

func TestAuthHandler_Register(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name               string
		requestBody        string
		mockRegisterFunc   func(ctx context.Context, login, password string) (string, error)
		expectedStatus     int
		expectedAuthHeader string
		checkAuthHeader    bool
	}{
		{
			name:        "Успешная регистрация",
			requestBody: `{"login":"testuser","password":"testpass"}`,
			mockRegisterFunc: func(ctx context.Context, login, password string) (string, error) {
				if login != "testuser" || password != "testpass" {
					t.Error("Ожидался login=testuser и password=testpass")
				}
				return "mock-jwt-token", nil
			},
			expectedStatus:     http.StatusOK,
			expectedAuthHeader: "Bearer mock-jwt-token",
			checkAuthHeader:    true,
		},
		{
			name:           "Невалидный JSON формат",
			requestBody:    `{"login":"test"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пустой логин",
			requestBody:    `{"login":"","password":"testpass"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пустой пароль",
			requestBody:    `{"login":"testuser","password":""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Оба поля пустые",
			requestBody:    `{"login":"","password":""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Пользователь уже существует",
			requestBody: `{"login":"existinguser","password":"testpass"}`,
			mockRegisterFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", service.ErrUserAlreadyExists
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "Внутренняя ошибка сервера",
			requestBody: `{"login":"testuser","password":"testpass"}`,
			mockRegisterFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				RegisterFunc: tt.mockRegisterFunc,
			}

			handler := NewAuthHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkAuthHeader {
				authHeader := rr.Header().Get("Authorization")
				if authHeader != tt.expectedAuthHeader {
					t.Errorf("Ожидался заголовок Authorization %q, получен %q", tt.expectedAuthHeader, authHeader)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name               string
		requestBody        string
		mockLoginFunc      func(ctx context.Context, login, password string) (string, error)
		expectedStatus     int
		expectedAuthHeader string
		checkAuthHeader    bool
	}{
		{
			name:        "Успешный логин",
			requestBody: `{"login":"testuser","password":"testpass"}`,
			mockLoginFunc: func(ctx context.Context, login, password string) (string, error) {
				if login != "testuser" || password != "testpass" {
					t.Error("Ожидался login=testuser и password=testpass")
				}
				return "mock-jwt-token-login", nil
			},
			expectedStatus:     http.StatusOK,
			expectedAuthHeader: "Bearer mock-jwt-token-login",
			checkAuthHeader:    true,
		},
		{
			name:           "Невалидный JSON формат",
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Неверные учетные данные",
			requestBody: `{"login":"testuser","password":"wrongpass"}`,
			mockLoginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", service.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "Пользователь не найден",
			requestBody: `{"login":"nonexistent","password":"testpass"}`,
			mockLoginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", service.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "Внутренняя ошибка сервера",
			requestBody: `{"login":"testuser","password":"testpass"}`,
			mockLoginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", errors.New("database connection error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				LoginFunc: tt.mockLoginFunc,
			}

			handler := NewAuthHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkAuthHeader {
				authHeader := rr.Header().Get("Authorization")
				if authHeader != tt.expectedAuthHeader {
					t.Errorf("Ожидался заголовок Authorization %q, получен %q", tt.expectedAuthHeader, authHeader)
				}
			}
		})
	}
}

func TestAuthHandler_Register_ContentType(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockAuthService{
		RegisterFunc: func(ctx context.Context, login, password string) (string, error) {
			return "token", nil
		},
	}
	handler := NewAuthHandler(logger, mockService)

	// Тест с application/json
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(`{"login":"user","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("С Content-Type application/json: ожидался статус %d, получен %d", http.StatusOK, rr.Code)
	}
}
