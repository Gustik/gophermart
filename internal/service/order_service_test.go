package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

func TestUploadOrder_Success(t *testing.T) {
	userID := 1
	orderNumber := "12345678903"

	mock := &mockStorage{
		getOrderByNumberFunc: func(ctx context.Context, number string) (*model.Order, error) {
			// Заказ не найден (новый заказ)
			return nil, storage.ErrOrderNotFound
		},
		createOrderFunc: func(ctx context.Context, userID int, number string) error {
			if number != orderNumber {
				t.Errorf("CreateOrder() number = %v, ожидалось %v", number, orderNumber)
			}
			return nil
		},
	}

	service := NewOrderService(mock)

	err := service.UploadOrder(context.Background(), userID, orderNumber)
	if err != nil {
		t.Errorf("UploadOrder() ошибка = %v, ожидалось nil", err)
	}
}

func TestUploadOrder_AlreadyUploadedBySameUser(t *testing.T) {
	userID := 1
	orderNumber := "12345678903"

	mock := &mockStorage{
		getOrderByNumberFunc: func(ctx context.Context, number string) (*model.Order, error) {
			// Заказ уже существует и принадлежит этому же пользователю
			return &model.Order{
				ID:         1,
				UserID:     userID,
				Number:     orderNumber,
				Status:     model.OrderStatusNew,
				UploadedAt: time.Now(),
			}, nil
		},
	}

	service := NewOrderService(mock)

	err := service.UploadOrder(context.Background(), userID, orderNumber)
	if !errors.Is(err, ErrOrderAlreadyUploaded) {
		t.Errorf("UploadOrder() ошибка = %v, ожидалось ErrOrderAlreadyUploaded", err)
	}
}

func TestUploadOrder_AlreadyUploadedByAnotherUser(t *testing.T) {
	userID := 1
	anotherUserID := 2
	orderNumber := "12345678903"

	mock := &mockStorage{
		getOrderByNumberFunc: func(ctx context.Context, number string) (*model.Order, error) {
			// Заказ уже существует и принадлежит другому пользователю
			return &model.Order{
				ID:         1,
				UserID:     anotherUserID,
				Number:     orderNumber,
				Status:     model.OrderStatusNew,
				UploadedAt: time.Now(),
			}, nil
		},
	}

	service := NewOrderService(mock)

	err := service.UploadOrder(context.Background(), userID, orderNumber)
	if !errors.Is(err, ErrOrderUploadedByAnotherUser) {
		t.Errorf("UploadOrder() ошибка = %v, ожидалось ErrOrderUploadedByAnotherUser", err)
	}
}

func TestUploadOrder_GetOrderByNumberError(t *testing.T) {
	userID := 1
	orderNumber := "12345678903"

	mock := &mockStorage{
		getOrderByNumberFunc: func(ctx context.Context, number string) (*model.Order, error) {
			// Ошибка базы данных (не ErrOrderNotFound)
			return nil, errors.New("database error")
		},
	}

	service := NewOrderService(mock)

	err := service.UploadOrder(context.Background(), userID, orderNumber)
	if err == nil {
		t.Error("UploadOrder() ожидалась ошибка при ошибке GetOrderByNumber")
	}
}

func TestUploadOrder_CreateOrderError(t *testing.T) {
	userID := 1
	orderNumber := "12345678903"

	mock := &mockStorage{
		getOrderByNumberFunc: func(ctx context.Context, number string) (*model.Order, error) {
			// Заказ не найден
			return nil, storage.ErrOrderNotFound
		},
		createOrderFunc: func(ctx context.Context, userID int, number string) error {
			// Ошибка при создании заказа
			return errors.New("database error")
		},
	}

	service := NewOrderService(mock)

	err := service.UploadOrder(context.Background(), userID, orderNumber)
	if err == nil {
		t.Error("UploadOrder() ожидалась ошибка при ошибке CreateOrder")
	}
}

func TestGetOrdersByUser_Success(t *testing.T) {
	userID := 1
	expectedOrders := []model.Order{
		{
			ID:         1,
			UserID:     userID,
			Number:     "12345678903",
			Status:     model.OrderStatusNew,
			UploadedAt: time.Now(),
		},
		{
			ID:         2,
			UserID:     userID,
			Number:     "79927398713",
			Status:     model.OrderStatusProcessing,
			UploadedAt: time.Now(),
		},
	}

	mock := &mockStorage{
		getOrdersByUserFunc: func(ctx context.Context, userID int) ([]model.Order, error) {
			return expectedOrders, nil
		},
	}

	service := NewOrderService(mock)

	orders, err := service.GetOrdersByUser(context.Background(), userID)
	if err != nil {
		t.Errorf("GetOrdersByUser() ошибка = %v, ожидалось nil", err)
		return
	}

	if len(orders) != len(expectedOrders) {
		t.Errorf("GetOrdersByUser() вернул %v заказов, ожидалось %v", len(orders), len(expectedOrders))
	}
}

func TestGetOrdersByUser_NoOrders(t *testing.T) {
	userID := 1

	mock := &mockStorage{
		getOrdersByUserFunc: func(ctx context.Context, userID int) ([]model.Order, error) {
			return []model.Order{}, nil
		},
	}

	service := NewOrderService(mock)

	orders, err := service.GetOrdersByUser(context.Background(), userID)
	if !errors.Is(err, ErrNoOrders) {
		t.Errorf("GetOrdersByUser() ошибка = %v, ожидалось ErrNoOrders", err)
	}
	if orders != nil {
		t.Error("GetOrdersByUser() должен вернуть nil при отсутствии заказов")
	}
}

func TestGetOrdersByUser_Error(t *testing.T) {
	userID := 1

	mock := &mockStorage{
		getOrdersByUserFunc: func(ctx context.Context, userID int) ([]model.Order, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewOrderService(mock)

	orders, err := service.GetOrdersByUser(context.Background(), userID)
	if err == nil {
		t.Error("GetOrdersByUser() ожидалась ошибка при ошибке GetOrdersByUser")
	}
	if orders != nil {
		t.Error("GetOrdersByUser() должен вернуть nil при ошибке")
	}
}
