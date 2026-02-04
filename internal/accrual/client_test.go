package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestGetOrderAccrual_Success(t *testing.T) {
	// Создаем mock сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/12345678903" {
			t.Errorf("Неожиданный путь: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345678903","status":"PROCESSED","accrual":500.5}`))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: server.URL,
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrderAccrual() ошибка = %v", err)
	}

	if result.Order != "12345678903" {
		t.Errorf("Order = %v, ожидалось 12345678903", result.Order)
	}
	if result.Status != AccrualStatusProcessed {
		t.Errorf("Status = %v, ожидалось PROCESSED", result.Status)
	}
	if result.Accrual == nil || *result.Accrual != 500.5 {
		t.Errorf("Accrual = %v, ожидалось 500.5", result.Accrual)
	}
}

func TestGetOrderAccrual_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: server.URL,
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != ErrOrderNotRegistered {
		t.Errorf("Ожидалась ошибка ErrOrderNotRegistered, получена %v", err)
	}
	if result != nil {
		t.Error("Result должен быть nil для 204 статуса")
	}
}

func TestGetOrderAccrual_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL:        server.URL,
		Logger:         logger,
		MaxRetries:     1,                    // Минимум retry
		InitialBackoff: 1 * time.Millisecond, // Быстрый backoff
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != ErrTooManyRequests {
		t.Errorf("Ожидалась ошибка ErrTooManyRequests, получена %v", err)
	}
	if result != nil {
		t.Error("Result должен быть nil для 429 статуса")
	}
}

func TestGetOrderAccrual_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL:        server.URL,
		Logger:         logger,
		MaxRetries:     1,                    // Минимум retry
		InitialBackoff: 1 * time.Millisecond, // Быстрый backoff
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err == nil {
		t.Error("Ожидалась ошибка для 500 статуса")
	}
	if result != nil {
		t.Error("Result должен быть nil при ошибке")
	}
}

func TestGetOrderAccrual_UnknowError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(505)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL:        server.URL,
		Logger:         logger,
		MaxRetries:     1,                    // Минимум retry
		InitialBackoff: 1 * time.Millisecond, // Быстрый backoff
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	expectedError := "неожиданный код ответа: 505"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка %q, получена %q", expectedError, err)
	}
	if result != nil {
		t.Error("Result должен быть nil при ошибке")
	}
}

func TestGetOrderAccrual_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345678903","status":"PROCESSED"`)) // невалидный JSON
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL:        server.URL,
		Logger:         logger,
		MaxRetries:     1,                    // Минимум retry
		InitialBackoff: 1 * time.Millisecond, // Быстрый backoff
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err == nil {
		t.Error("Ожидалась ошибка при невалидном JSON")
	}
	if result != nil {
		t.Error("Result должен быть nil при ошибке парсинга")
	}
}

func TestGetOrderAccrual_StatusRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345678903","status":"REGISTERED"}`))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: server.URL,
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrderAccrual() ошибка = %v", err)
	}

	if result.Status != AccrualStatusRegistered {
		t.Errorf("Status = %v, ожидалось REGISTERED", result.Status)
	}
}

func TestGetOrderAccrual_StatusProcessing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345678903","status":"PROCESSING"}`))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: server.URL,
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrderAccrual() ошибка = %v", err)
	}

	if result.Status != AccrualStatusProcessing {
		t.Errorf("Status = %v, ожидалось PROCESSING", result.Status)
	}
}

func TestGetOrderAccrual_StatusInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345678903","status":"INVALID"}`))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: server.URL,
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrderAccrual() ошибка = %v", err)
	}

	if result.Status != AccrualStatusInvalid {
		t.Errorf("Status = %v, ожидалось INVALID", result.Status)
	}
}

func TestNewClient_DefaultValues(t *testing.T) {
	logger := zap.NewNop()
	client, err := NewClient(Config{
		BaseURL: "http://localhost:8080",
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("Не удалось создать клиент: %v", err)
	}

	if client == nil {
		t.Error("Client не должен быть nil")
	}
}

func TestNewClient_NoLogger(t *testing.T) {
	_, err := NewClient(Config{
		BaseURL: "http://localhost:8080",
		Logger:  nil,
	})
	if err == nil {
		t.Error("Ожидалась ошибка при отсутствии логгера")
	}
}

func TestParseRetryAfter_Seconds(t *testing.T) {
	logger := zap.NewNop()
	c := &client{
		logger:         logger,
		initialBackoff: 1 * time.Second,
	}

	duration := c.parseRetryAfter("5")
	if duration != 5*time.Second {
		t.Errorf("parseRetryAfter(\"5\") = %v, ожидалось 5s", duration)
	}
}

func TestParseRetryAfter_Empty(t *testing.T) {
	logger := zap.NewNop()
	initialBackoff := 2 * time.Second
	c := &client{
		logger:         logger,
		initialBackoff: initialBackoff,
	}

	duration := c.parseRetryAfter("")
	if duration != initialBackoff {
		t.Errorf("parseRetryAfter(\"\") = %v, ожидалось %v", duration, initialBackoff)
	}
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	logger := zap.NewNop()
	c := &client{
		logger:         logger,
		initialBackoff: 1 * time.Second,
	}

	// Время в будущем (UTC для корректного форматирования)
	future := time.Now().UTC().Add(10 * time.Second)
	httpDate := future.Format(http.TimeFormat)

	duration := c.parseRetryAfter(httpDate)
	// Проверяем, что duration примерно равно 10 секундам (с погрешностью)
	if duration < 9*time.Second || duration > 11*time.Second {
		t.Errorf("parseRetryAfter(future) = %v, ожидалось ~10s", duration)
	}
}

func TestParseRetryAfter_InvalidFormat(t *testing.T) {
	logger := zap.NewNop()
	initialBackoff := 3 * time.Second
	c := &client{
		logger:         logger,
		initialBackoff: initialBackoff,
	}

	duration := c.parseRetryAfter("invalid")
	if duration != initialBackoff {
		t.Errorf("parseRetryAfter(\"invalid\") = %v, ожидалось %v", duration, initialBackoff)
	}
}
