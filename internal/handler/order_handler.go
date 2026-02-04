package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Gustik/gophermart/internal/auth"
	"github.com/Gustik/gophermart/internal/service"
	"github.com/Gustik/gophermart/internal/validator"
	"go.uber.org/zap"
)

type OrderHandler struct {
	*BaseHandler
	orderService service.OrderService
}

// NewOrderHandler создает новый OrderHandler
func NewOrderHandler(logger *zap.Logger, orderService service.OrderService) *OrderHandler {
	handler := &OrderHandler{
		BaseHandler:  NewBaseHandler(logger),
		orderService: orderService,
	}

	handler.errorStatusMap = map[error]int{
		validator.ErrEmptyOrderNumber:         http.StatusBadRequest,
		validator.ErrInvalidOrderFormat:       http.StatusUnprocessableEntity,
		service.ErrOrderAlreadyUploaded:       http.StatusOK,
		service.ErrOrderUploadedByAnotherUser: http.StatusConflict,
		service.ErrNoOrders:                   http.StatusNoContent,
	}

	return handler
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
		h.HandleError(w, err)
		return
	}

	err = h.orderService.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
