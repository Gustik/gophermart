package service

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/accrual"
	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

// mockAccrualClient - мок для accrual клиента
type mockAccrualClient struct {
	getOrderAccrualFunc func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error)
}

func (m *mockAccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
	if m.getOrderAccrualFunc != nil {
		return m.getOrderAccrualFunc(ctx, orderNumber)
	}
	return nil, errors.New("not implemented")
}

func TestProcessOrder_Registered(t *testing.T) {
	orderNumber := "12345678903"
	accrualValue := float32(500.0)

	mock := &mockStorage{
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			if status != model.OrderStatusProcessing {
				t.Errorf("ожидался статус PROCESSING, получен %v", status)
			}
			if accrual != nil {
				t.Errorf("для статуса REGISTERED не должно быть accrual")
			}
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:   orderNumber,
				Status:  accrual.AccrualStatusRegistered,
				Accrual: &accrualValue,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil", err)
	}
}

func TestProcessOrder_Processing(t *testing.T) {
	orderNumber := "12345678903"

	mock := &mockStorage{
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			t.Errorf("не должен вызываться UpdateOrderStatus для статуса PROCESSING")
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:  orderNumber,
				Status: accrual.AccrualStatusProcessing,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil", err)
	}
}

func TestProcessOrder_Invalid(t *testing.T) {
	orderNumber := "12345678903"

	mock := &mockStorage{
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			if status != model.OrderStatusInvalid {
				t.Errorf("ожидался статус INVALID, получен %v", status)
			}
			if accrual != nil {
				t.Errorf("для статуса INVALID не должно быть accrual")
			}
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:  orderNumber,
				Status: accrual.AccrualStatusInvalid,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil", err)
	}
}

func TestProcessOrder_Processed(t *testing.T) {
	orderNumber := "12345678903"
	accrualValue := float32(500.0)
	updateStatusCalled := false
	addBalanceCalled := false

	mock := &mockStorage{
		beginTxFunc: func(ctx context.Context) (storage.Tx, error) {
			return &mockTx{}, nil
		},
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			updateStatusCalled = true
			if status != model.OrderStatusProcessed {
				t.Errorf("ожидался статус PROCESSED, получен %v", status)
			}
			if accrual == nil || *accrual != accrualValue {
				t.Errorf("ожидалось accrual = %v, получено %v", accrualValue, accrual)
			}
			return nil
		},
		addBalanceFunc: func(ctx context.Context, tx storage.Tx, userID int, amount *float32) error {
			addBalanceCalled = true
			if amount == nil || *amount != accrualValue {
				t.Errorf("ожидалось amount = %v, получено %v", accrualValue, amount)
			}
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:   orderNumber,
				Status:  accrual.AccrualStatusProcessed,
				Accrual: &accrualValue,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil", err)
	}

	if !updateStatusCalled {
		t.Error("UpdateOrderStatus не был вызван")
	}
	if !addBalanceCalled {
		t.Error("AddBalance не был вызван")
	}
}

func TestProcessOrder_ProcessedWithZeroAccrual(t *testing.T) {
	orderNumber := "12345678903"
	accrualValue := float32(0)
	updateStatusCalled := false

	mock := &mockStorage{
		beginTxFunc: func(ctx context.Context) (storage.Tx, error) {
			return &mockTx{}, nil
		},
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			updateStatusCalled = true
			if status != model.OrderStatusProcessed {
				t.Errorf("ожидался статус PROCESSED, получен %v", status)
			}
			return nil
		},
		addBalanceFunc: func(ctx context.Context, tx storage.Tx, userID int, amount *float32) error {
			t.Errorf("AddBalance не должен вызываться при нулевом начислении")
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:   orderNumber,
				Status:  accrual.AccrualStatusProcessed,
				Accrual: &accrualValue,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil", err)
	}

	if !updateStatusCalled {
		t.Error("UpdateOrderStatus не был вызван")
	}
}

func TestProcessOrder_OrderNotRegistered(t *testing.T) {
	orderNumber := "12345678903"

	mock := &mockStorage{
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			t.Errorf("не должен вызываться UpdateOrderStatus для незарегистрированного заказа")
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return nil, accrual.ErrOrderNotRegistered
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err != nil {
		t.Errorf("processOrder() ошибка = %v, ожидалось nil (не ошибка, пропускаем)", err)
	}
}

func TestProcessOrder_TooManyRequests(t *testing.T) {
	orderNumber := "12345678903"

	mock := &mockStorage{}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return nil, accrual.ErrTooManyRequests
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err == nil {
		t.Error("processOrder() ожидалась ошибка для TooManyRequests")
	}
}

func TestProcessOrder_ServerError(t *testing.T) {
	orderNumber := "12345678903"

	mock := &mockStorage{}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return nil, accrual.ErrServerError
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	err := service.processOrder(context.Background(), order)
	if err == nil {
		t.Error("processOrder() ожидалась ошибка для ServerError")
	}
}

func TestProcessPendingOrders(t *testing.T) {
	accrualValue := float32(500.0)

	orders := []model.Order{
		{ID: 1, UserID: 1, Number: "12345678903", Status: model.OrderStatusNew},
		{ID: 2, UserID: 1, Number: "79927398713", Status: model.OrderStatusProcessing},
	}

	processedOrders := make(map[string]bool)

	mock := &mockStorage{
		getPendingOrdersFunc: func(ctx context.Context) ([]model.Order, error) {
			return orders, nil
		},
		beginTxFunc: func(ctx context.Context) (storage.Tx, error) {
			return &mockTx{}, nil
		},
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			processedOrders[orderNumber] = true
			return nil
		},
		addBalanceFunc: func(ctx context.Context, tx storage.Tx, userID int, amount *float32) error {
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			return &accrual.OrderAccrual{
				Order:   orderNumber,
				Status:  accrual.AccrualStatusProcessed,
				Accrual: &accrualValue,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	err := service.ProcessPendingOrders(context.Background())
	if err != nil {
		t.Errorf("ProcessPendingOrders() ошибка = %v, ожидалось nil", err)
	}

	// Проверяем, что оба заказа были обработаны
	if !processedOrders["12345678903"] {
		t.Error("заказ 12345678903 не был обработан")
	}
	if !processedOrders["79927398713"] {
		t.Error("заказ 79927398713 не был обработан")
	}
}

func TestProcessPendingOrders_GetPendingOrdersError(t *testing.T) {
	mock := &mockStorage{
		getPendingOrdersFunc: func(ctx context.Context) ([]model.Order, error) {
			return nil, errors.New("database error")
		},
	}

	mockClient := &mockAccrualClient{}
	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	err := service.ProcessPendingOrders(context.Background())
	if err == nil {
		t.Error("ProcessPendingOrders() ожидалась ошибка при ошибке получения заказов")
	}
}

func TestProcessPendingOrders_ContinuesOnError(t *testing.T) {
	orders := []model.Order{
		{ID: 1, UserID: 1, Number: "12345678903", Status: model.OrderStatusNew},
		{ID: 2, UserID: 1, Number: "79927398713", Status: model.OrderStatusNew},
	}

	processedOrders := make(map[string]bool)

	mock := &mockStorage{
		getPendingOrdersFunc: func(ctx context.Context) ([]model.Order, error) {
			return orders, nil
		},
		updateOrderStatusFunc: func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
			processedOrders[orderNumber] = true
			return nil
		},
	}

	mockClient := &mockAccrualClient{
		getOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*accrual.OrderAccrual, error) {
			// Первый заказ возвращает ошибку, второй - успех
			if orderNumber == "12345678903" {
				return nil, errors.New("some error")
			}
			return &accrual.OrderAccrual{
				Order:  orderNumber,
				Status: accrual.AccrualStatusRegistered,
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewAccrualService(mockClient, mock, logger)

	err := service.ProcessPendingOrders(context.Background())
	if err != nil {
		t.Errorf("ProcessPendingOrders() ошибка = %v, ожидалось nil", err)
	}

	// Проверяем, что второй заказ был обработан несмотря на ошибку первого
	if !processedOrders["79927398713"] {
		t.Error("заказ 79927398713 не был обработан, хотя должен был продолжить после ошибки")
	}
}
