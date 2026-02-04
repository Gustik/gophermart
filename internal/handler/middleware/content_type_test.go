package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireContentType(t *testing.T) {
	// Создаем тестовый обработчик, который просто возвращает 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		contentType    string
		requiredType   string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Правильный Content-Type",
			contentType:    "application/json",
			requiredType:   "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Неправильный Content-Type",
			contentType:    "text/plain",
			requiredType:   "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Неверный Content-Type\n",
		},
		{
			name:           "Отсутствует Content-Type",
			contentType:    "",
			requiredType:   "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Неверный Content-Type\n",
		},
		{
			name:           "Content-Type text/plain требуется text/plain",
			contentType:    "text/plain",
			requiredType:   "text/plain",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Content-Type с дополнительными параметрами",
			contentType:    "application/json; charset=utf-8",
			requiredType:   "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Неверный Content-Type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Применяем middleware
			middleware := RequireContentType(tt.requiredType)
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			// Проверяем тело ответа
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Ожидалось тело %q, получено %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestRequireContentType_MethodChaining(t *testing.T) {
	// Тестируем, что middleware можно комбинировать
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Применяем два middleware подряд
	middleware1 := RequireContentType("application/json")
	middleware2 := RequireContentType("application/json")
	handler := middleware1(middleware2(testHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, rr.Code)
	}
}
