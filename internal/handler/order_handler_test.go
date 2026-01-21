package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
)

// MockOrderService для тестирования
type MockOrderService struct {
	UploadOrderFunc func(ctx context.Context, userID int, orderNumber string) error
}

func (m *MockOrderService) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	if m.UploadOrderFunc != nil {
		return m.UploadOrderFunc(ctx, userID, orderNumber)
	}
	return nil
}

func TestOrderHandler_UploadOrder(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name              string
		requestBody       string
		userID            int
		setupContext      bool
		mockUploadFunc    func(ctx context.Context, userID int, orderNumber string) error
		expectedStatus    int
		expectedBodyPart  string
		checkBodyContains bool
	}{
		{
			name:         "Успешная загрузка - новый заказ",
			requestBody:  "12345678903",
			userID:       123,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				if userID != 123 {
					t.Errorf("Ожидался userID=123, получен %d", userID)
				}
				if orderNumber != "12345678903" {
					t.Errorf("Ожидался orderNumber=12345678903, получен %s", orderNumber)
				}
				return nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:         "Заказ уже загружен этим пользователем",
			requestBody:  "12345678903",
			userID:       123,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrOrderAlreadyUploaded
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "Заказ загружен другим пользователем",
			requestBody:  "12345678903",
			userID:       123,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrOrderUploadedByAnotherUser
			},
			expectedStatus:    http.StatusConflict,
			expectedBodyPart:  "заказ уже был загружен другим пользователем",
			checkBodyContains: true,
		},
		{
			name:              "Отсутствует контекст авторизации",
			requestBody:       "12345678903",
			setupContext:      false,
			expectedStatus:    http.StatusUnauthorized,
			expectedBodyPart:  "Не авторизован",
			checkBodyContains: true,
		},
		{
			name:              "Пустой номер заказа",
			requestBody:       "",
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusBadRequest,
			expectedBodyPart:  "номер заказа не может быть пустым",
			checkBodyContains: true,
		},
		{
			name:              "Неверный формат заказа - неправильная контрольная сумма",
			requestBody:       "12345678901",
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedBodyPart:  "неверный формат номера заказа",
			checkBodyContains: true,
		},
		{
			name:              "Неверный формат заказа - содержит буквы",
			requestBody:       "1234567890a",
			userID:            123,
			setupContext:      true,
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedBodyPart:  "неверный формат номера заказа",
			checkBodyContains: true,
		},
		{
			name:         "Заказ с пробелами",
			requestBody:  "  12345678903  ",
			userID:       123,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				if orderNumber != "12345678903" {
					t.Errorf("Ожидался обрезанный orderNumber=12345678903, получен %s", orderNumber)
				}
				return nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:         "Внутренняя ошибка сервера",
			requestBody:  "12345678903",
			userID:       123,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return errors.New("database connection error")
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedBodyPart:  "Внутренняя ошибка сервера",
			checkBodyContains: true,
		},
		{
			name:         "Валидный номер по Луну - 79927398713",
			requestBody:  "79927398713",
			userID:       456,
			setupContext: true,
			mockUploadFunc: func(ctx context.Context, userID int, orderNumber string) error {
				if orderNumber != "79927398713" {
					t.Errorf("Ожидался orderNumber=79927398713, получен %s", orderNumber)
				}
				return nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:              "Невалидный номер по Луну - 79927398710",
			requestBody:       "79927398710",
			userID:            456,
			setupContext:      true,
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedBodyPart:  "неверный формат номера заказа",
			checkBodyContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockOrderService{
				UploadOrderFunc: tt.mockUploadFunc,
			}

			handler := NewOrderHandler(logger, mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "text/plain")

			// Устанавливаем userID в контекст, если требуется
			if tt.setupContext {
				ctx := auth.SetUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.UploadOrder(rr, req)

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

func TestOrderHandler_UploadOrder_ReadError(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockOrderService{}
	handler := NewOrderHandler(logger, mockService)

	// Создаем request с телом, которое вызовет ошибку при чтении
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", errReader{})
	ctx := auth.SetUserID(req.Context(), 123)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UploadOrder(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d для ошибки чтения, получен %d", http.StatusBadRequest, rr.Code)
	}
}

// errReader реализует io.Reader, который всегда возвращает ошибку
type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// Вспомогательная функция для проверки содержимого строки
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(str) > 0 && len(substr) > 0 && indexOf(str, substr) >= 0))
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
