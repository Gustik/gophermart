package accrual

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	// ErrOrderNotRegistered возвращается когда заказ не зарегистрирован в системе
	ErrOrderNotRegistered = errors.New("заказ не зарегистрирован в системе начисления")
	// ErrTooManyRequests возвращается при превышении лимита запросов
	ErrTooManyRequests = errors.New("превышен лимит запросов к системе начисления")
	// ErrServerError возвращается при внутренней ошибке сервера
	ErrServerError = errors.New("внутренняя ошибка системы начисления")
)

// AccrualStatus представляет статус расчёта начисления
type AccrualStatus string

const (
	// AccrualStatusRegistered - заказ зарегистрирован, но начисление не рассчитано
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	// AccrualStatusInvalid - заказ не принят к расчёту
	AccrualStatusInvalid AccrualStatus = "INVALID"
	// AccrualStatusProcessing - расчёт начисления в процессе
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	// AccrualStatusProcessed - расчёт начисления окончен
	AccrualStatusProcessed AccrualStatus = "PROCESSED"
)

// OrderAccrual представляет информацию о начислении баллов за заказ
type OrderAccrual struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float32      `json:"accrual,omitempty"`
}

// Client определяет интерфейс клиента для системы начисления
type Client interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*OrderAccrual, error)
}

// client представляет HTTP-клиент для взаимодействия с системой начисления
type client struct {
	client         *resty.Client
	logger         *zap.Logger
	initialBackoff time.Duration
	retryAfter     time.Duration // динамическое значение из заголовка Retry-After
}

// Config содержит конфигурацию клиента
type Config struct {
	BaseURL         string
	Logger          *zap.Logger
	Timeout         time.Duration
	MaxRetries      int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiple float64
}

// NewClient создаёт новый клиент для системы начисления
func NewClient(cfg Config) (Client, error) {
	// Значения по умолчанию
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.InitialBackoff == 0 {
		cfg.InitialBackoff = 1 * time.Second
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	if cfg.BackoffMultiple == 0 {
		cfg.BackoffMultiple = 2.0
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("не передан логгер")
	}

	// Создаём клиент (понадобится для замыкания в хуках)
	accrualClient := &client{
		logger:         cfg.Logger,
		initialBackoff: cfg.InitialBackoff,
	}

	// Создаём resty клиент
	client := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.MaxRetries).
		SetRetryWaitTime(cfg.InitialBackoff).
		SetRetryMaxWaitTime(cfg.MaxBackoff).
		SetLogger(&zapLoggerAdapter{logger: cfg.Logger})

	// Настраиваем retry условия
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Повторяем при ошибках сети или 5xx ошибках
		if err != nil {
			return true
		}
		// Не повторяем при 204 (заказ не найден)
		if r.StatusCode() == http.StatusNoContent {
			return false
		}
		// Повторяем при 429 и 500
		return r.StatusCode() == http.StatusTooManyRequests ||
			r.StatusCode() == http.StatusInternalServerError
	})

	// Обработка Retry-After заголовка
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// Парсим Retry-After при 429 и сохраняем для следующего retry
		if resp.StatusCode() == http.StatusTooManyRequests {
			retryAfter := accrualClient.parseRetryAfter(resp.Header().Get("Retry-After"))
			accrualClient.retryAfter = retryAfter

			accrualClient.logger.Warn("получен Retry-After заголовок",
				zap.Duration("повтор_через", retryAfter),
			)
		}
		return nil
	})

	// Кастомная задержка для Retry-After
	client.AddRetryHook(func(r *resty.Response, err error) {
		// Если есть сохранённое значение Retry-After, используем его
		if accrualClient.retryAfter > 0 {
			accrualClient.logger.Warn("ожидание перед повтором согласно Retry-After",
				zap.Duration("задержка", accrualClient.retryAfter),
			)
			time.Sleep(accrualClient.retryAfter)
			accrualClient.retryAfter = 0 // сбрасываем после использования
		}
	})

	accrualClient.client = client
	return accrualClient, nil
}

// parseRetryAfter парсит заголовок Retry-After
func (c *client) parseRetryAfter(header string) time.Duration {
	if header == "" {
		return c.initialBackoff
	}

	// Пробуем распарсить как количество секунд
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Пробуем распарсить как HTTP-дату
	if t, err := http.ParseTime(header); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return c.initialBackoff
}
