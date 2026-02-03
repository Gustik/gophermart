package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/service"
)

// MockBalanceService для тестирования
type MockBalanceService struct {
	GetBalanceFunc     func(ctx context.Context, userID int) (*model.Balance, error)
	WithdrawFunc       func(ctx context.Context, userID int, orderNumber string, sum float32) error
	GetWithdrawalsFunc func(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

func (m *MockBalanceService) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockBalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float32) error {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, userID, orderNumber, sum)
	}
	return nil
}

func (m *MockBalanceService) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	if m.GetWithdrawalsFunc != nil {
		return m.GetWithdrawalsFunc(ctx, userID)
	}
	return nil, nil
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name              string
		userID            int
		setupContext      bool
		mockGetBalanceFunc func(ctx context.Context, userID int) (*model.Balance, error)
		expectedStatus    int
		expectedBodyPart  string
		checkBodyContains bool
		checkJSON         bool
	}{
		{
			name:         "Успешное получение баланса",
			userID:       123,
			setupContext: true,
			mockGetBalanceFunc: func(ctx context.Context, userID int) (*model.Balance, error) {
				if userID != 123 {
					t.Errorf("Ожидался userID=123, получен %d", userID)
				}
				return &model.Balance{
					UserID:    123,
					Current:   500.5,
					Withdrawn: 42.0,
					UpdatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:              "Отсутствует контекст авторизации",
			setupContext:      false,
			expectedStatus:    http.StatusUnauthorized,
			expectedBodyPart:  "Не авторизован",
			checkBodyContains: true,
		},
		{
			name:         "Внутренняя ошибка сервера",
			userID:       123,
			setupContext: true,
			mockGetBalanceFunc: func(ctx context.Context, userID int) (*model.Balance, error) {
				return nil, errors.New("database connection error")
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedBodyPart:  "Внутренняя ошибка сервера",
			checkBodyContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockBalanceService{
				GetBalanceFunc: tt.mockGetBalanceFunc,
			}

			handler := NewBalanceHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

			// Устанавливаем userID в контекст, если требуется
			if tt.setupContext {
				ctx := auth.SetUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.GetBalance(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d. Тело: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.checkBodyContains {
				body := rr.Body.String()
				if !contains(body, tt.expectedBodyPart) {
					t.Errorf("Ожидалось, что тело содержит %q, получено %q", tt.expectedBodyPart, body)
				}
			}

			if tt.checkJSON {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Ожидался Content-Type application/json, получен %s", contentType)
				}

				body := rr.Body.String()
				if body == "" {
					t.Error("Ожидался непустой JSON ответ")
				}
			}
		})
	}
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name              string
		requestBody       string
		userID            int
		setupContext      bool
		mockWithdrawFunc  func(ctx context.Context, userID int, orderNumber string, sum float32) error
		expectedStatus    int
		expectedBodyPart  string
		checkBodyContains bool
	}{
		{
			name:         "Успешное списание",
			requestBody:  `{"order":"12345678903","sum":100.5}`,
			userID:       123,
			setupContext: true,
			mockWithdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float32) error {
				if userID != 123 {
					t.Errorf("Ожидался userID=123, получен %d", userID)
				}
				if orderNumber != "12345678903" {
					t.Errorf("Ожидался orderNumber=12345678903, получен %s", orderNumber)
				}
				if sum != 100.5 {
					t.Errorf("Ожидалась sum=100.5, получена %f", sum)
				}
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "Недостаточно средств",
			requestBody:  `{"order":"12345678903","sum":1000.0}`,
			userID:       123,
			setupContext: true,
			mockWithdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float32) error {
				return service.ErrInsufficientBalance
			},
			expectedStatus:    http.StatusPaymentRequired,
			expectedBodyPart:  "недостаточно средств",
			checkBodyContains: true,
		},
		{
			name:              "Отсутствует контекст авторизации",
			requestBody:       `{"order":"12345678903","sum":100.0}`,
			setupContext:      false,
			expectedStatus:    http.StatusUnauthorized,
			expectedBodyPart:  "Не авторизован",
			checkBodyContains: true,
		},
		{
			name:              "Невалидный JSON",
			requestBody:       `{"order":"12345678903"`,
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusBadRequest,
			expectedBodyPart:  "Неверный формат запроса",
			checkBodyContains: true,
		},
		{
			name:              "Неверный номер заказа - не проходит Луна",
			requestBody:       `{"order":"12345678901","sum":100.0}`,
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedBodyPart:  "Неверный номер заказа",
			checkBodyContains: true,
		},
		{
			name:              "Пустой номер заказа",
			requestBody:       `{"order":"","sum":100.0}`,
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedBodyPart:  "Неверный номер заказа",
			checkBodyContains: true,
		},
		{
			name:         "Внутренняя ошибка сервера",
			requestBody:  `{"order":"12345678903","sum":100.0}`,
			userID:       123,
			setupContext: true,
			mockWithdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float32) error {
				return errors.New("database error")
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedBodyPart:  "Внутренняя ошибка сервера",
			checkBodyContains: true,
		},
		{
			name:         "Валидный номер по Луну - 79927398713",
			requestBody:  `{"order":"79927398713","sum":50.0}`,
			userID:       456,
			setupContext: true,
			mockWithdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float32) error {
				if orderNumber != "79927398713" {
					t.Errorf("Ожидался orderNumber=79927398713, получен %s", orderNumber)
				}
				return nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockBalanceService{
				WithdrawFunc: tt.mockWithdrawFunc,
			}

			handler := NewBalanceHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Устанавливаем userID в контекст, если требуется
			if tt.setupContext {
				ctx := auth.SetUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.Withdraw(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d. Тело: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.checkBodyContains {
				body := rr.Body.String()
				if !contains(body, tt.expectedBodyPart) {
					t.Errorf("Ожидалось, что тело содержит %q, получено %q", tt.expectedBodyPart, body)
				}
			}
		})
	}
}

func TestBalanceHandler_Withdraw_ReadError(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockBalanceService{}
	handler := NewBalanceHandler(logger, mockService)

	// Создаем request с телом, которое вызовет ошибку при чтении
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", errReader{})
	ctx := auth.SetUserID(req.Context(), 123)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Withdraw(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d для ошибки чтения, получен %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBalanceHandler_GetWithdrawals(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name                   string
		userID                 int
		setupContext           bool
		mockGetWithdrawalsFunc func(ctx context.Context, userID int) ([]model.Withdrawal, error)
		expectedStatus         int
		expectedBodyPart       string
		checkBodyContains      bool
		checkJSON              bool
	}{
		{
			name:         "Успешное получение списания",
			userID:       123,
			setupContext: true,
			mockGetWithdrawalsFunc: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				if userID != 123 {
					t.Errorf("Ожидался userID=123, получен %d", userID)
				}
				return []model.Withdrawal{
					{
						ID:          1,
						UserID:      123,
						OrderNumber: "12345678903",
						Sum:         100.5,
						ProcessedAt: time.Now(),
					},
					{
						ID:          2,
						UserID:      123,
						OrderNumber: "79927398713",
						Sum:         42.0,
						ProcessedAt: time.Now(),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:         "Нет списаний",
			userID:       123,
			setupContext: true,
			mockGetWithdrawalsFunc: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				return nil, service.ErrNoWithdrawals
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:              "Отсутствует контекст авторизации",
			setupContext:      false,
			expectedStatus:    http.StatusUnauthorized,
			expectedBodyPart:  "Не авторизован",
			checkBodyContains: true,
		},
		{
			name:         "Внутренняя ошибка сервера",
			userID:       123,
			setupContext: true,
			mockGetWithdrawalsFunc: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				return nil, errors.New("database connection error")
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedBodyPart:  "Внутренняя ошибка сервера",
			checkBodyContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockBalanceService{
				GetWithdrawalsFunc: tt.mockGetWithdrawalsFunc,
			}

			handler := NewBalanceHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)

			// Устанавливаем userID в контекст, если требуется
			if tt.setupContext {
				ctx := auth.SetUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.GetWithdrawals(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d. Тело: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.checkBodyContains {
				body := rr.Body.String()
				if !contains(body, tt.expectedBodyPart) {
					t.Errorf("Ожидалось, что тело содержит %q, получено %q", tt.expectedBodyPart, body)
				}
			}

			if tt.checkJSON {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Ожидался Content-Type application/json, получен %s", contentType)
				}

				body := rr.Body.String()
				if body == "" {
					t.Error("Ожидался непустой JSON ответ")
				}
			}
		})
	}
}
