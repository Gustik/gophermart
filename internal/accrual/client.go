package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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
	Order   string         `json:"order"`
	Status  AccrualStatus  `json:"status"`
	Accrual *float32       `json:"accrual,omitempty"`
}

// Client представляет HTTP-клиент для взаимодействия с системой начисления
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	// Параметры retry логики
	maxRetries      int
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	backoffMultiple float64
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
func NewClient(cfg Config) *Client {
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

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger:          cfg.Logger,
		maxRetries:      cfg.MaxRetries,
		initialBackoff:  cfg.InitialBackoff,
		maxBackoff:      cfg.MaxBackoff,
		backoffMultiple: cfg.BackoffMultiple,
	}
}

// GetOrderAccrual получает информацию о начислении баллов за заказ
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*OrderAccrual, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	var lastErr error
	backoff := c.initialBackoff

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Info("повторный запрос в систему начисления",
				zap.String("заказ", orderNumber),
				zap.Int("попытка", attempt),
				zap.Duration("задержка", backoff),
			)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			// Увеличиваем backoff с exponential backoff
			backoff = time.Duration(float64(backoff) * c.backoffMultiple)
			if backoff > c.maxBackoff {
				backoff = c.maxBackoff
			}
		}

		accrual, retryAfter, err := c.doRequest(ctx, url, orderNumber)
		if err == nil {
			return accrual, nil
		}

		lastErr = err

		// Если получили 429 с Retry-After, используем это значение
		if errors.Is(err, ErrTooManyRequests) && retryAfter > 0 {
			c.logger.Warn("превышен лимит запросов, повтор через",
				zap.String("заказ", orderNumber),
				zap.Duration("повтор_через", retryAfter),
			)
			backoff = retryAfter
			continue
		}

		// Если заказ не найден, не делаем retry
		if errors.Is(err, ErrOrderNotRegistered) {
			return nil, err
		}

		// Для других ошибок продолжаем retry
		c.logger.Warn("ошибка запроса в систему начисления",
			zap.String("заказ", orderNumber),
			zap.Error(err),
			zap.Int("попытка", attempt),
		)
	}

	c.logger.Error("исчерпаны все попытки повтора",
		zap.String("заказ", orderNumber),
		zap.Error(lastErr),
	)

	return nil, lastErr
}

// doRequest выполняет HTTP-запрос к системе начисления
func (c *Client) doRequest(ctx context.Context, url, orderNumber string) (*OrderAccrual, time.Duration, error) {
	c.logger.Info("отправка запроса в систему начисления",
		zap.String("url", url),
		zap.String("заказ", orderNumber),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось отправить запрос: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Info("получен ответ от системы начисления",
		zap.String("заказ", orderNumber),
		zap.Int("статус_код", resp.StatusCode),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, 0, fmt.Errorf("не удалось прочитать тело ответа: %w", err)
		}

		var accrual OrderAccrual
		if err := json.Unmarshal(body, &accrual); err != nil {
			return nil, 0, fmt.Errorf("не удалось распарсить ответ: %w", err)
		}

		c.logger.Info("успешно распарсен ответ системы начисления",
			zap.String("заказ", orderNumber),
			zap.String("статус", string(accrual.Status)),
			zap.Any("начисление", accrual.Accrual),
		)

		return &accrual, 0, nil

	case http.StatusNoContent:
		c.logger.Info("заказ не зарегистрирован в системе начисления",
			zap.String("заказ", orderNumber),
		)
		return nil, 0, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := c.parseRetryAfter(resp.Header.Get("Retry-After"))
		c.logger.Warn("превышен лимит запросов",
			zap.String("заказ", orderNumber),
			zap.Duration("повтор_через", retryAfter),
		)
		return nil, retryAfter, ErrTooManyRequests

	case http.StatusInternalServerError:
		c.logger.Error("внутренняя ошибка системы начисления",
			zap.String("заказ", orderNumber),
		)
		return nil, 0, ErrServerError

	default:
		c.logger.Error("неожиданный код ответа от системы начисления",
			zap.String("заказ", orderNumber),
			zap.Int("статус_код", resp.StatusCode),
		)
		return nil, 0, fmt.Errorf("неожиданный код ответа: %d", resp.StatusCode)
	}
}

// parseRetryAfter парсит заголовок Retry-After
func (c *Client) parseRetryAfter(header string) time.Duration {
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
