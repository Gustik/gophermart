package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
	"github.com/Gustik/gophermart/internal/validator"
	"go.uber.org/zap"
)

// withdrawRequest представляет запрос на списание баллов
type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

// BalanceHandler обрабатывает HTTP-запросы связанные с балансом пользователя
type BalanceHandler struct {
	*BaseHandler
	balanceService service.BalanceService
}

// NewBalanceHandler создаёт новый BalanceHandler
func NewBalanceHandler(logger *zap.Logger, balanceService service.BalanceService) *BalanceHandler {
	handler := &BalanceHandler{
		BaseHandler:    NewBaseHandler(logger),
		balanceService: balanceService,
	}

	handler.errorStatusMap = map[error]int{
		service.ErrInsufficientBalance: http.StatusPaymentRequired, // 402
		service.ErrNoWithdrawals:       http.StatusNoContent,       // 204
	}

	return handler
}

// GetBalance возвращает текущий баланс пользователя
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

// Withdraw обрабатывает запрос на списание баллов с баланса
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Парсим JSON
	var req withdrawRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Валидируем номер заказа по алгоритму Луна
	if err := validator.ValidateOrderNumber(req.Order); err != nil {
		http.Error(w, "Неверный номер заказа", http.StatusUnprocessableEntity)
		return
	}

	// Списываем баллы
	err = h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientBalance) {
			http.Error(w, "На счету недостаточно средств", http.StatusPaymentRequired)
			return
		}
		h.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals возвращает историю списаний пользователя
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNoWithdrawals) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
