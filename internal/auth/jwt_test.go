package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := 123
	login := "testuser"

	token, err := GenerateToken(secret, userID, login)
	if err != nil {
		t.Fatalf("GenerateToken() ошибка = %v, ожидалось nil", err)
	}

	if token == "" {
		t.Error("GenerateToken() вернул пустой токен")
	}

	// Проверяем, что токен можно распарсить
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("Не удалось распарсить токен: %v", err)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		t.Fatal("Не удалось получить claims из токена")
	}

	if claims.UserID != userID {
		t.Errorf("Claims.UserID = %v, ожидалось %v", claims.UserID, userID)
	}

	if claims.Login != login {
		t.Errorf("Claims.Login = %v, ожидалось %v", claims.Login, login)
	}

	// Проверяем, что срок действия токена установлен
	if claims.ExpiresAt == nil {
		t.Error("Claims.ExpiresAt не установлен")
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := 456
	login := "validuser"

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   bool
		wantUser  int
		wantLogin string
	}{
		{
			name: "Валидный токен",
			token: func() string {
				token, _ := GenerateToken(secret, userID, login)
				return token
			}(),
			secret:    secret,
			wantErr:   false,
			wantUser:  userID,
			wantLogin: login,
		},
		{
			name:    "Неверный secret",
			token: func() string {
				token, _ := GenerateToken(secret, userID, login)
				return token
			}(),
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "Невалидный токен",
			token:   "invalid.token.string",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "Пустой токен",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Истекший токен",
			token: func() string {
				// Создаем токен с истекшим сроком
				claims := &Claims{
					UserID: userID,
					Login:  login,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			}(),
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.secret, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateToken() ожидалась ошибка, но её не было")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() ошибка = %v, ожидалось nil", err)
				return
			}

			if claims.UserID != tt.wantUser {
				t.Errorf("Claims.UserID = %v, ожидалось %v", claims.UserID, tt.wantUser)
			}

			if claims.Login != tt.wantLogin {
				t.Errorf("Claims.Login = %v, ожидалось %v", claims.Login, tt.wantLogin)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() http.Request
		wantUserID int
		wantOk     bool
	}{
		{
			name: "UserID присутствует в контексте",
			setupCtx: func() http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Используем auth.SetUserID, а проверяем auth.GetUserID
				ctx := SetUserID(req.Context(), 789)
				return *req.WithContext(ctx)
			},
			wantUserID: 789,
			wantOk:     true,
		},
		{
			name: "UserID отсутствует в контексте",
			setupCtx: func() http.Request {
				return *httptest.NewRequest(http.MethodGet, "/", nil)
			},
			wantUserID: 0,
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupCtx()
			gotUserID, gotOk := GetUserID(req.Context())

			if gotOk != tt.wantOk {
				t.Errorf("GetUserID() ok = %v, ожидалось %v", gotOk, tt.wantOk)
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("GetUserID() userID = %v, ожидалось %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	// Тестируем с пустым секретом
	token, err := GenerateToken("", 1, "testuser")
	if err != nil {
		t.Errorf("GenerateToken с пустым secret не должен возвращать ошибку, получена: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken с пустым secret должен вернуть токен")
	}
}

func TestSetUserID_MultipleValues(t *testing.T) {
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	// Устанавливаем первое значение
	ctx = SetUserID(ctx, 100)
	userID, ok := GetUserID(ctx)
	if !ok || userID != 100 {
		t.Errorf("Первое значение: userID = %v, ok = %v, ожидалось 100, true", userID, ok)
	}

	// Перезаписываем значение
	ctx = SetUserID(ctx, 200)
	userID, ok = GetUserID(ctx)
	if !ok || userID != 200 {
		t.Errorf("Второе значение: userID = %v, ok = %v, ожидалось 200, true", userID, ok)
	}
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	secret := "test-secret"

	// Генерируем токены для разных пользователей
	token1, err1 := GenerateToken(secret, 1, "user1")
	token2, err2 := GenerateToken(secret, 2, "user2")

	if err1 != nil || err2 != nil {
		t.Fatal("Ошибка при генерации токенов")
	}

	if token1 == token2 {
		t.Error("Токены для разных пользователей должны быть разными")
	}

	// Проверяем, что можем валидировать оба токена
	claims1, err := ValidateToken(secret, token1)
	if err != nil || claims1.UserID != 1 {
		t.Error("Токен пользователя 1 должен быть валидным")
	}

	claims2, err := ValidateToken(secret, token2)
	if err != nil || claims2.UserID != 2 {
		t.Error("Токен пользователя 2 должен быть валидным")
	}
}
