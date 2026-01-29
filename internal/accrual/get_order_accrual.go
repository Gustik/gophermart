package accrual

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// GetOrderAccrual получает информацию о начислении баллов за заказ
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*OrderAccrual, error) {
	var result OrderAccrual

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		SetPathParam("number", orderNumber).
		Get("/api/orders/{number}")

	if err != nil {
		c.logger.Error("не удалось выполнить запрос",
			zap.String("order", orderNumber),
			zap.Error(err),
		)
		return nil, fmt.Errorf("не удалось отправить запрос: %w", err)
	}

	// Обрабатываем различные статус-коды
	switch resp.StatusCode() {
	case http.StatusOK:
		c.logger.Info("успешно получена информация о начислении",
			zap.String("order", orderNumber),
			zap.String("status", string(result.Status)),
			zap.Any("accrual", result.Accrual),
		)
		return &result, nil

	case http.StatusNoContent:
		c.logger.Info("заказ не зарегистрирован в системе начисления",
			zap.String("order", orderNumber),
		)
		return nil, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		// Retry логика с Retry-After обрабатывается автоматически через retry hooks
		c.logger.Error("превышен лимит запросов после всех retry попыток",
			zap.String("order", orderNumber),
		)
		return nil, ErrTooManyRequests

	case http.StatusInternalServerError:
		c.logger.Error("внутренняя ошибка системы начисления",
			zap.String("order", orderNumber),
		)
		return nil, ErrServerError

	default:
		c.logger.Error("неожиданный код ответа от системы начисления",
			zap.String("order", orderNumber),
			zap.Int("status", resp.StatusCode()),
		)
		return nil, fmt.Errorf("неожиданный код ответа: %d", resp.StatusCode())
	}
}
