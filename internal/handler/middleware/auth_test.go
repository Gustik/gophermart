package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Gustik/gophermart/internal/auth"
)

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := "test-secret-key-12345"

	// Создаем тестовый обработчик
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetUserID(r.Context())
		if !ok {
			t.Error("Expected userID in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userID != 123 {
			t.Errorf("Expected userID=123, got %d", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Валидный токен",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Отсутствует заголовок Authorization",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Отсутствует токен авторизации\n",
		},
		{
			name:           "Неверный формат - нет префикса Bearer",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Неверный формат токена\n",
		},
		{
			name:           "Неверный формат - только Bearer",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Неверный формат токена\n",
		},
		{
			name:           "Невалидная подпись токена",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный токен\n",
		},
		{
			name:           "Токен подписан неверным секретом",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный токен\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Для валидного токена генерируем настоящий JWT
			authHeader := tt.authHeader
			if tt.name == "Валидный токен" {
				token, err := auth.GenerateToken(jwtSecret, 123, "testuser")
				if err != nil {
					t.Fatalf("Не удалось сгенерировать токен: %v", err)
				}
				authHeader = "Bearer " + token
			} else if tt.name == "Токен подписан неверным секретом" {
				// Генерируем токен с другим секретом
				wrongSecret := "wrong-secret"
				token, err := auth.GenerateToken(wrongSecret, 123, "testuser")
				if err != nil {
					t.Fatalf("Не удалось сгенерировать токен: %v", err)
				}
				authHeader = "Bearer " + token
			}

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Применяем middleware
			middleware := AuthMiddleware(jwtSecret)
			handler := middleware(testHandler)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			// Проверяем тело ответа (для ошибок)
			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("Ожидалось тело %q, получено %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestAuthMiddleware_ContextPropagation(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := 456
	login := "testuser"

	// Генерируем валидный токен
	token, err := auth.GenerateToken(jwtSecret, userID, login)
	if err != nil {
		t.Fatalf("Не удалось сгенерировать токен: %v", err)
	}

	// Создаем обработчик, который проверяет контекст
	var contextUserID int
	var contextOk bool
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserID, contextOk = auth.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Создаем запрос с валидным токеном
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	// Применяем middleware
	middleware := AuthMiddleware(jwtSecret)
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем, что userID правильно передан в контекст
	if !contextOk {
		t.Error("Ожидалось наличие userID в контексте")
	}
	if contextUserID != userID {
		t.Errorf("Ожидался userID=%d в контексте, получен %d", userID, contextUserID)
	}
}
