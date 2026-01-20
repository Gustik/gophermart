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

// OrderService предоставляет бизнес-логику для заказов
type OrderService struct {
	storage *storage.Storage
}

func NewOrderService(storage *storage.Storage) *OrderService {
	return &OrderService{
		storage: storage,
	}
}

// UploadOrder загружает номер заказа
func (s *OrderService) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	order, err := s.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
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
