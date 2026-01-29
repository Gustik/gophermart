package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/service"
)

type AccrualWorker struct {
	service      *service.AccrualService
	logger       *zap.Logger
	pollInterval time.Duration // как часто опрашивать
	batchSize    int           // сколько заказов обрабатывать за раз
	stopCh       chan struct{} // для graceful shutdown
	doneCh       chan struct{} // сигнал что worker остановлен
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewAccrualWorker(service *service.AccrualService, logger *zap.Logger, pollInterval time.Duration, batchSize int) *AccrualWorker {
	return &AccrualWorker{
		service:      service,
		logger:       logger,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
}

// Start запускает worker в отдельной горутине
func (w *AccrualWorker) Start() {
	w.ctx, w.cancel = context.WithCancel(context.Background())
	go w.run()
}

// Stop останавливает worker gracefully
func (w *AccrualWorker) Stop() {
	w.logger.Info("Остановка воркера...")
	w.cancel()
	close(w.stopCh)
	<-w.doneCh // ждём завершения
	w.logger.Info("Воркер остановлен")
}

// run - основной цикл
func (w *AccrualWorker) run() {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info("Воркер запущен")

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.service.ProcessPendingOrders(w.ctx); err != nil {
				w.logger.Error("ошибка обработки заказов", zap.Error(err))
			}
		}
	}
}
