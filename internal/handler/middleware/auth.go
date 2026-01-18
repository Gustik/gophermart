// internal/handlers/middleware/auth.go

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Gustik/gophermart/internal/auth"
)

// Ключ для контекста
type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из заголовка
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Отсутствует токен авторизации", http.StatusUnauthorized)
				return
			}

			// Проверяем формат "Bearer TOKEN"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Валидируем токен
			claims, err := auth.ValidateToken(jwtSecret, tokenString)
			if err != nil {
				http.Error(w, "Невалидный токен", http.StatusUnauthorized)
				return
			}

			// Добавляем user_id в контекст
			ctx := auth.SetUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID извлекает user_id из контекста
func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}
