package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/accrual"
	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
)

type AccrualService struct {
	client  accrual.Client
	storage storage.Storage
	logger  *zap.Logger
}

func NewAccrualService(client accrual.Client, storage storage.Storage, logger *zap.Logger) *AccrualService {
	return &AccrualService{
		client:  client,
		storage: storage,
		logger:  logger,
	}
}

func (s *AccrualService) ProcessPendingOrders(ctx context.Context) error {
	orders, err := s.storage.GetPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить заказы для обработки: %w", err)
	}

	for i := range orders {
		if err := s.processOrder(ctx, &orders[i]); err != nil {
			s.logger.Error("не удалось обработать заказ",
				zap.String("order", orders[i].Number),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *AccrualService) processOrder(ctx context.Context, order *model.Order) error {
	acc, err := s.client.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if errors.Is(err, accrual.ErrOrderNotRegistered) {
			return nil // не ошибка, просто пропускаем
		}
		return fmt.Errorf("не удалось получить статус: %w", err)
	}

	switch acc.Status {
	case accrual.AccrualStatusRegistered:
		return s.storage.UpdateOrderStatus(ctx, nil, order.Number, model.OrderStatusProcessing, nil)

	case accrual.AccrualStatusProcessing:
		return nil // ждём дальнейшей обработки

	case accrual.AccrualStatusInvalid:
		return s.storage.UpdateOrderStatus(ctx, nil, order.Number, model.OrderStatusInvalid, nil)

	case accrual.AccrualStatusProcessed:
		return s.processCompletedOrder(ctx, order, acc)
	}

	return nil
}

func (s *AccrualService) processCompletedOrder(ctx context.Context, order *model.Order, acc *accrual.OrderAccrual) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	if err := s.storage.UpdateOrderStatus(ctx, tx, order.Number, model.OrderStatusProcessed, acc.Accrual); err != nil {
		return fmt.Errorf("не удалось обновить статус: %w", err)
	}

	if acc.Accrual != nil && *acc.Accrual > 0 {
		if err := s.storage.AddBalance(ctx, tx, order.UserID, acc.Accrual); err != nil {
			return fmt.Errorf("не удалось обновить баланс: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return nil
}
