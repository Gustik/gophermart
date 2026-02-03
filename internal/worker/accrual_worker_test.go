package worker

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/service"
)

func TestNewAccrualWorker(t *testing.T) {
	logger := zap.NewNop()
	pollInterval := 5 * time.Second
	batchSize := 10

	// Создаем AccrualService с nil зависимостями для теста
	accrualService := service.NewAccrualService(nil, nil, logger)
	worker := NewAccrualWorker(accrualService, logger, pollInterval, batchSize)

	if worker == nil {
		t.Fatal("NewAccrualWorker() вернул nil")
	}

	if worker.pollInterval != pollInterval {
		t.Errorf("pollInterval = %v, ожидалось %v", worker.pollInterval, pollInterval)
	}

	if worker.batchSize != batchSize {
		t.Errorf("batchSize = %v, ожидалось %v", worker.batchSize, batchSize)
	}

	if worker.logger == nil {
		t.Error("logger не должен быть nil")
	}

	if worker.stopCh == nil {
		t.Error("stopCh не должен быть nil")
	}

	if worker.doneCh == nil {
		t.Error("doneCh не должен быть nil")
	}
}

func TestAccrualWorker_StartStop(t *testing.T) {
	logger := zap.NewNop()
	accrualService := service.NewAccrualService(nil, nil, logger)

	// Используем большой интервал, чтобы ticker не успел сработать
	worker := NewAccrualWorker(accrualService, logger, 10*time.Second, 10)

	// Запускаем worker
	worker.Start()

	// Быстро останавливаем worker
	worker.Stop()

	// Проверяем, что worker установил контекст
	if worker.ctx == nil {
		t.Error("ctx должен быть установлен после Start()")
	}

	if worker.cancel == nil {
		t.Error("cancel должен быть установлен после Start()")
	}
}

func TestAccrualWorker_StopWithoutStart(t *testing.T) {
	logger := zap.NewNop()
	accrualService := service.NewAccrualService(nil, nil, logger)
	worker := NewAccrualWorker(accrualService, logger, 1*time.Second, 10)

	// Пытаемся остановить worker, который не был запущен
	// Это вызовет panic, так как cancel == nil
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Ожидаем panic
				done <- true
			} else {
				done <- false
			}
		}()
		worker.Stop()
	}()

	select {
	case result := <-done:
		if !result {
			t.Error("Ожидалась паника при вызове Stop() без Start()")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Stop() завис")
	}
}

func TestAccrualWorker_QuickStartStop(t *testing.T) {
	logger := zap.NewNop()
	accrualService := service.NewAccrualService(nil, nil, logger)
	worker := NewAccrualWorker(accrualService, logger, 10*time.Millisecond, 10)

	// Быстрый запуск и остановка
	worker.Start()
	worker.Stop()

	// Проверяем, что worker корректно завершился
	if worker.ctx.Err() == nil {
		t.Error("ctx должен быть отменен после Stop()")
	}
}

func TestAccrualWorker_MultipleStartStop(t *testing.T) {
	logger := zap.NewNop()
	accrualService := service.NewAccrualService(nil, nil, logger)

	// Используем большой интервал, чтобы ticker не успел сработать
	// и не вызвал ProcessPendingOrders с nil зависимостями
	worker1 := NewAccrualWorker(accrualService, logger, 10*time.Second, 10)
	worker1.Start()
	worker1.Stop()

	// Второй worker
	worker2 := NewAccrualWorker(accrualService, logger, 10*time.Second, 10)
	worker2.Start()
	worker2.Stop()
}

func TestAccrualWorker_DifferentIntervals(t *testing.T) {
	logger := zap.NewNop()
	accrualService := service.NewAccrualService(nil, nil, logger)

	tests := []struct {
		name         string
		pollInterval time.Duration
		batchSize    int
	}{
		{"Short interval", 1 * time.Millisecond, 5},
		{"Medium interval", 100 * time.Millisecond, 10},
		{"Long interval", 1 * time.Second, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := NewAccrualWorker(accrualService, logger, tt.pollInterval, tt.batchSize)

			if worker.pollInterval != tt.pollInterval {
				t.Errorf("pollInterval = %v, ожидалось %v", worker.pollInterval, tt.pollInterval)
			}

			if worker.batchSize != tt.batchSize {
				t.Errorf("batchSize = %v, ожидалось %v", worker.batchSize, tt.batchSize)
			}
		})
	}
}
