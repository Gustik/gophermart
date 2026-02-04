package service

import (
	"context"
	"errors"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

var (
	ErrOrderAlreadyUploaded       = errors.New("заказ уже был загружен этим пользователем")
	ErrOrderUploadedByAnotherUser = errors.New("заказ уже был загружен другим пользователем")
	ErrNoOrders                   = errors.New("не заказов")
)

// OrderService определяет методы для сервиса заказов
type OrderService interface {
	UploadOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error)
}

// orderService предоставляет бизнес-логику для заказов
type orderService struct {
	storage storage.Storage
}

// NewOrderService создаёт новый OrderService
func NewOrderService(storage storage.Storage) OrderService {
	return &orderService{
		storage: storage,
	}
}

// UploadOrder загружает номер заказа
func (s *orderService) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	order, err := s.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, storage.ErrOrderNotFound) {
		return err
	}

	if order != nil {
		if order.UserID != userID {
			return ErrOrderUploadedByAnotherUser
		}

		return ErrOrderAlreadyUploaded
	}

	err = s.storage.CreateOrder(ctx, userID, orderNumber)
	if err != nil {
		return err
	}

	return nil
}

// GetOrdersByUser возвращает заказы пользователя
func (s *orderService) GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error) {
	orders, err := s.storage.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrNoOrders
	}

	return orders, nil
}
