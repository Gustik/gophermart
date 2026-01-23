package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
	"go.uber.org/zap"
)

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	BaseHandler
	authService service.AuthService
}

// NewAuthHandler создаёт новый AuthHandler
func NewAuthHandler(logger *zap.Logger, authService service.AuthService) *AuthHandler {
	handler := &AuthHandler{
		BaseHandler: *NewBaseHandler(logger),
		authService: authService,
	}

	handler.errorStatusMap = map[error]int{
		service.ErrUserAlreadyExists:  http.StatusConflict,
		service.ErrInvalidCredentials: http.StatusUnauthorized,
	}

	return handler
}

// RegisterRequest - запрос на регистрацию
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		h.BaseHandler.HandleError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login аутентификация пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.BaseHandler.HandleError(w, err)
		return
	}

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
