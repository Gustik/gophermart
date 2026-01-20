package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
	"github.com/Gustik/gophermart/internal/validator"
	"go.uber.org/zap"
)

type OrderHandler struct {
	logger       *zap.Logger
	orderService *service.OrderService
}

// NewOrderHandler создает новый OrderHandler
func NewOrderHandler(logger *zap.Logger, orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		logger:       logger,
		orderService: orderService,
	}
}

// UploadOrder загрузка номера заказа
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))

	if err := validator.ValidateOrderNumber(orderNumber); err != nil {
		if errors.Is(err, validator.ErrEmptyOrderNumber) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, validator.ErrInvalidOrderFormat) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}

	err = h.orderService.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, service.ErrOrderAlreadyUploaded) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, service.ErrOrderUploadedByAnotherUser) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
