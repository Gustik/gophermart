package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

var (
	// ErrInsufficientBalance возвращается при недостаточном балансе для списания (HTTP 402)
	ErrInsufficientBalance = errors.New("недостаточно средств на балансе")
	// ErrNoWithdrawals возвращается когда нет истории списаний (HTTP 204)
	ErrNoWithdrawals = errors.New("нет операций списания")
)

// BalanceService определяет методы для сервиса баланса
type BalanceService interface {
	// GetBalance возвращает текущий баланс пользователя (current и withdrawn)
	GetBalance(ctx context.Context, userID int) (*model.Balance, error)
	// Withdraw списывает баллы с баланса пользователя в счёт оплаты заказа
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float32) error
	// GetWithdrawals возвращает историю списаний пользователя, отсортированную от новых к старым
	GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

// balanceService предоставляет бизнес-логику для работы с балансом
type balanceService struct {
	storage storage.Storage
}

// NewBalanceService создаёт новый BalanceService
func NewBalanceService(storage storage.Storage) BalanceService {
	return &balanceService{
		storage: storage,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (s *balanceService) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	balance, err := s.storage.GetBalance(ctx, nil, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	return balance, nil
}

// Withdraw списывает баллы с баланса пользователя
func (s *balanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float32) error {
	// Начинаем транзакцию для атомарности операции
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	// Получаем текущий баланс с блокировкой для обновления
	balance, err := s.storage.GetBalance(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения баланса: %w", err)
	}

	// Проверяем достаточность средств
	if balance.Current < sum {
		return ErrInsufficientBalance
	}

	// Списываем средства с баланса
	if err := s.storage.WithdrawBalance(ctx, tx, userID, sum); err != nil {
		return fmt.Errorf("ошибка списания баланса: %w", err)
	}

	// Создаём запись о списании
	if err := s.storage.CreateWithdrawal(ctx, tx, userID, orderNumber, sum); err != nil {
		return fmt.Errorf("ошибка создания записи о списании: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (s *balanceService) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	withdrawals, err := s.storage.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списаний: %w", err)
	}

	if len(withdrawals) == 0 {
		return nil, ErrNoWithdrawals
	}

	return withdrawals, nil
}
