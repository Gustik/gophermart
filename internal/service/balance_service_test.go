package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

func TestGetBalance(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		mockSetup func(*mockStorage)
		want      *model.Balance
		wantErr   bool
	}{
		{
			name:   "Успешное получение баланса",
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{
						UserID:    1,
						Current:   100.5,
						Withdrawn: 50.0,
						UpdatedAt: time.Now(),
					}, nil
				}
			},
			want: &model.Balance{
				UserID:    1,
				Current:   100.5,
				Withdrawn: 50.0,
			},
			wantErr: false,
		},
		{
			name:   "Ошибка при получении баланса",
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return nil, errors.New("database error")
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			tt.mockSetup(mock)

			s := NewBalanceService(mock)

			got, err := s.GetBalance(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() ошибка = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Current != tt.want.Current || got.Withdrawn != tt.want.Withdrawn) {
				t.Errorf("GetBalance() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		orderNumber string
		sum         float32
		mockSetup   func(*mockStorage)
		wantErr     error
	}{
		{
			name:        "Успешное списание",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{Current: 100.0, Withdrawn: 0}, nil
				}
				m.withdrawBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int, sum float32) error {
					return nil
				}
				m.createWithdrawalFunc = func(ctx context.Context, tx storage.Tx, userID int, orderNumber string, sum float32) error {
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name:        "Недостаточно средств",
			userID:      1,
			orderNumber: "12345678903",
			sum:         150.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{Current: 100.0}, nil
				}
			},
			wantErr: ErrInsufficientBalance,
		},
		{
			name:        "Ошибка начала транзакции",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return nil, errors.New("transaction error")
				}
			},
			wantErr: nil, // проверяем, что есть ошибка
		},
		{
			name:        "Ошибка получения баланса",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr: nil, // проверяем, что есть ошибка
		},
		{
			name:        "Ошибка списания баланса",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{Current: 100.0}, nil
				}
				m.withdrawBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int, sum float32) error {
					return errors.New("withdraw error")
				}
			},
			wantErr: nil, // проверяем, что есть ошибка
		},
		{
			name:        "Ошибка создания записи о списании",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{Current: 100.0}, nil
				}
				m.withdrawBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int, sum float32) error {
					return nil
				}
				m.createWithdrawalFunc = func(ctx context.Context, tx storage.Tx, userID int, orderNumber string, sum float32) error {
					return errors.New("create withdrawal error")
				}
			},
			wantErr: nil, // проверяем, что есть ошибка
		},
		{
			name:        "Ошибка коммита транзакции",
			userID:      1,
			orderNumber: "12345678903",
			sum:         50.0,
			mockSetup: func(m *mockStorage) {
				m.beginTxFunc = func(ctx context.Context) (storage.Tx, error) {
					return &mockTx{
						commitFunc: func() error {
							return errors.New("commit error")
						},
					}, nil
				}
				m.getBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int) (*model.Balance, error) {
					return &model.Balance{Current: 100.0}, nil
				}
				m.withdrawBalanceFunc = func(ctx context.Context, tx storage.Tx, userID int, sum float32) error {
					return nil
				}
				m.createWithdrawalFunc = func(ctx context.Context, tx storage.Tx, userID int, orderNumber string, sum float32) error {
					return nil
				}
			},
			wantErr: nil, // проверяем, что есть ошибка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			tt.mockSetup(mock)

			s := NewBalanceService(mock)

			err := s.Withdraw(context.Background(), tt.userID, tt.orderNumber, tt.sum)

			if tt.wantErr != nil {
				// Проверяем конкретную ошибку
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Withdraw() ошибка = %v, wantErr %v", err, tt.wantErr)
				}
			} else if tt.name == "Успешное списание" {
				// Для успешного случая не должно быть ошибки
				if err != nil {
					t.Errorf("Withdraw() ошибка = %v, ожидалось nil", err)
				}
			} else {
				// Для остальных случаев просто проверяем наличие ошибки
				if err == nil {
					t.Errorf("Withdraw() ожидалась ошибка, но её не было")
				}
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		mockSetup func(*mockStorage)
		wantLen   int
		wantErr   error
	}{
		{
			name:   "Успешное получение списаний",
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.getWithdrawalsFunc = func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
					return []model.Withdrawal{
						{OrderNumber: "12345678903", Sum: 50.0},
						{OrderNumber: "79927398713", Sum: 25.0},
					}, nil
				}
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name:   "Нет списаний",
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.getWithdrawalsFunc = func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
					return []model.Withdrawal{}, nil
				}
			},
			wantLen: 0,
			wantErr: ErrNoWithdrawals,
		},
		{
			name:   "Ошибка получения списаний",
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.getWithdrawalsFunc = func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
					return nil, errors.New("database error")
				}
			},
			wantLen: 0,
			wantErr: nil, // проверяем, что есть ошибка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			tt.mockSetup(mock)

			s := NewBalanceService(mock)

			got, err := s.GetWithdrawals(context.Background(), tt.userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetWithdrawals() ошибка = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if tt.name == "Успешное получение списаний" {
				if err != nil {
					t.Errorf("GetWithdrawals() ошибка = %v, ожидалось nil", err)
					return
				}
			} else {
				// Для случаев с ошибками проверяем, что ошибка есть
				if err == nil {
					t.Errorf("GetWithdrawals() ожидалась ошибка, но её не было")
					return
				}
			}

			if err == nil && len(got) != tt.wantLen {
				t.Errorf("GetWithdrawals() = %v элементов, want %v", len(got), tt.wantLen)
			}
		})
	}
}
