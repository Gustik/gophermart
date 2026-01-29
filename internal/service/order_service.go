package service

import (
	"context"
	"errors"

	"github.com/Gustik/gophermart/internal/storage"
)

var (
	ErrOrderAlreadyUploaded       = errors.New("заказ уже был загружен этим пользователем")
	ErrOrderUploadedByAnotherUser = errors.New("заказ уже был загружен другим пользователем")
)

// OrderService определяет методы для сервиса заказов
type OrderService interface {
	UploadOrder(ctx context.Context, userID int, orderNumber string) error
}

// orderService предоставляет бизнес-логику для заказов
type orderService struct {
	storage *storage.Storage
}

// NewOrderService создаёт новый OrderService
func NewOrderService(storage *storage.Storage) OrderService {
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
			return ErrOrderAlreadyUploaded
		}

		return ErrOrderAlreadyUploaded
	}

	err = s.storage.CreateOrder(ctx, userID, orderNumber)
	if err != nil {
		return err
	}

	return nil
}
