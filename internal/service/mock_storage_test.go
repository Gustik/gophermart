package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

// mockTx - мок для транзакции
type mockTx struct {
	rollbackFunc func() error
	commitFunc   func() error
}

func (m *mockTx) Rollback() error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc()
	}
	return nil
}

func (m *mockTx) Commit() error {
	if m.commitFunc != nil {
		return m.commitFunc()
	}
	return nil
}

// mockStorage - мок для storage
type mockStorage struct {
	getBalanceFunc        func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error)
	beginTxFunc           func(ctx context.Context) (storage.Tx, error)
	withdrawBalanceFunc   func(ctx context.Context, tx storage.Tx, userID int, sum float32) error
	createWithdrawalFunc  func(ctx context.Context, tx storage.Tx, userID int, orderNumber string, sum float32) error
	getWithdrawalsFunc    func(ctx context.Context, userID int) ([]model.Withdrawal, error)
	getPendingOrdersFunc  func(ctx context.Context) ([]model.Order, error)
	updateOrderStatusFunc func(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error
	addBalanceFunc        func(ctx context.Context, tx storage.Tx, userID int, amount *float32) error
	getUserByLoginFunc    func(ctx context.Context, login string) (*model.User, error)
	createUserFunc        func(ctx context.Context, login, passwordHash string) (*model.User, error)
	initBalanceFunc       func(ctx context.Context, userID int) error
}

func (m *mockStorage) GetBalance(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(ctx, tx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockStorage) BeginTx(ctx context.Context) (storage.Tx, error) {
	if m.beginTxFunc != nil {
		return m.beginTxFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockStorage) WithdrawBalance(ctx context.Context, tx storage.Tx, userID int, sum float32) error {
	if m.withdrawBalanceFunc != nil {
		return m.withdrawBalanceFunc(ctx, tx, userID, sum)
	}
	return errors.New("not implemented")
}

func (m *mockStorage) CreateWithdrawal(ctx context.Context, tx storage.Tx, userID int, orderNumber string, sum float32) error {
	if m.createWithdrawalFunc != nil {
		return m.createWithdrawalFunc(ctx, tx, userID, orderNumber, sum)
	}
	return errors.New("not implemented")
}

func (m *mockStorage) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	if m.getWithdrawalsFunc != nil {
		return m.getWithdrawalsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

// Остальные методы интерфейса (не используются в тестах)
func (m *mockStorage) Close() error                              { return nil }
func (m *mockStorage) RunMigrations(migrationsPath string) error { return nil }
func (m *mockStorage) DB() *sqlx.DB                              { return nil }
func (m *mockStorage) CreateUser(ctx context.Context, login, passwordHash string) (*model.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, login, passwordHash)
	}
	return nil, errors.New("not implemented")
}
func (m *mockStorage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if m.getUserByLoginFunc != nil {
		return m.getUserByLoginFunc(ctx, login)
	}
	return nil, errors.New("not implemented")
}
func (m *mockStorage) CreateOrder(ctx context.Context, userID int, number string) error {
	return errors.New("not implemented")
}
func (m *mockStorage) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) GetPendingOrders(ctx context.Context) ([]model.Order, error) {
	if m.getPendingOrdersFunc != nil {
		return m.getPendingOrdersFunc(ctx)
	}
	return nil, errors.New("not implemented")
}
func (m *mockStorage) UpdateOrderStatus(ctx context.Context, tx storage.Tx, orderNumber string, status model.OrderStatus, accrual *float32) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, tx, orderNumber, status, accrual)
	}
	return errors.New("not implemented")
}
func (m *mockStorage) InitBalance(ctx context.Context, userID int) error {
	if m.initBalanceFunc != nil {
		return m.initBalanceFunc(ctx, userID)
	}
	return errors.New("not implemented")
}
func (m *mockStorage) AddBalance(ctx context.Context, tx storage.Tx, userID int, amount *float32) error {
	if m.addBalanceFunc != nil {
		return m.addBalanceFunc(ctx, tx, userID, amount)
	}
	return errors.New("not implemented")
}
