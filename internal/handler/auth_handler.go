package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
	"go.uber.org/zap"
)

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	logger      *zap.Logger
	authService *service.AuthService
}

// NewAuthHandler создаёт новый AuthHandler
func NewAuthHandler(logger *zap.Logger, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		logger:      logger,
		authService: authService,
	}
}

// RegisterRequest - запрос на регистрацию
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register обрабатывает POST /api/user/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Парсим запрос
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	// Вызываем бизнес-логику
	token, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			http.Error(w, "Логин уже занят", http.StatusConflict)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает POST /api/user/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Парсим запрос
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Вызываем бизнес-логику
	token, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "Неверная пара логин/пароль", http.StatusUnauthorized)
			return
		}

		h.logger.Sugar().Errorf("Внутрення ошибка %w", err)

		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Check хендлер для проверки авторизации
func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		w.Write([]byte("no auth"))
		return
	}

	w.Write([]byte(strconv.Itoa(userID)))
}
