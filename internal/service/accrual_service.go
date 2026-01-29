package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gustik/gophermart/internal/accrual"
	"github.com/Gustik/gophermart/internal/model"
	"github.com/Gustik/gophermart/internal/storage"
	"go.uber.org/zap"
)

type AccrualService struct {
	client  *accrual.Client
	storage *storage.Storage
	logger  *zap.Logger
}

func NewAccrualService(client *accrual.Client, storage *storage.Storage, logger *zap.Logger) *AccrualService {
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
		acc, err := s.client.GetOrderAccrual(ctx, orders[i].Number)
		if err != nil {
			if errors.Is(err, accrual.ErrOrderNotRegistered) {
				continue
			}

			s.logger.Error("не удалось получить статус заказа",
				zap.String("order", orders[i].Number),
				zap.Error(err),
			)
			continue
		}

		switch acc.Status {
		case accrual.AccrualStatusRegistered:
			if err = s.storage.UpdateOrderStatus(ctx, orders[i].Number, model.OrderStatusProcessing, nil); err != nil {
				s.logger.Error("не удалось обновить статус заказа",
					zap.String("order", orders[i].Number),
					zap.Error(err),
				)
				continue
			}
		case accrual.AccrualStatusProcessing:
			continue
		case accrual.AccrualStatusInvalid:
			if err = s.storage.UpdateOrderStatus(ctx, orders[i].Number, model.OrderStatusInvalid, nil); err != nil {
				s.logger.Error("не удалось обновить статус заказа",
					zap.String("order", orders[i].Number),
					zap.Error(err),
				)
				continue
			}
		case accrual.AccrualStatusProcessed:
			if err = s.storage.UpdateOrderStatus(ctx, orders[i].Number, model.OrderStatusProcessed, acc.Accrual); err != nil {
				s.logger.Error("не удалось обновить статус заказа",
					zap.String("order", orders[i].Number),
					zap.Error(err),
				)
				continue
			}
		}
	}

	return nil
}
