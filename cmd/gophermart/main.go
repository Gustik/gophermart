// cmd/gophermart/main.go

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Gustik/gophermart/internal/accrual"
	"github.com/Gustik/gophermart/internal/config"
	"github.com/Gustik/gophermart/internal/handler"
	"github.com/Gustik/gophermart/internal/service"
	"github.com/Gustik/gophermart/internal/storage"
	"github.com/Gustik/gophermart/internal/worker"
	"github.com/Gustik/gophermart/internal/zaplog"
)

func main() {
	cfg := config.New()

	logger := mustInitLogger()
	defer logger.Sync()

	printStartupInfo(logger, *cfg)

	store := mustInitStorage(logger, cfg.DatabaseURI)
	defer store.Close()

	server := startServer(store, logger, cfg.RunAddress, cfg.JWTSecret)
	accrualWorker := startAccrualWorker(logger, store, cfg.AccrualSystemAddress)

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	accrualWorker.Stop()

	logger.Info("Остановка сервера...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Sugar().Fatalf("Принудительная остановка сервера: %v", err)
	}

	logger.Info("Сервер остановлен")
}

// maskPassword маскирует пароль в URI для безопасного логирования
func maskPassword(uri string) string {
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	return re.ReplaceAllString(uri, "://$1:***@")
}

func printStartupInfo(logger *zap.Logger, cfg config.Config) {
	logger.Info("Запуск сервера gophermart")
	logger.Sugar().Infof("Адрес сервера: %s", cfg.RunAddress)
	logger.Sugar().Infof("База данных: %s", maskPassword(cfg.DatabaseURI))
	logger.Sugar().Infof("Система начисления: %s", cfg.AccrualSystemAddress)
}

func mustInitLogger() *zap.Logger {
	logger, err := zaplog.New("info")
	if err != nil {
		logger.Fatal("Не удалось подключиться к БД", zap.Error(err))
	}
	return logger
}

func mustInitStorage(logger *zap.Logger, databaseURI string) *storage.Storage {
	store, err := storage.New(databaseURI)
	if err != nil {
		logger.Fatal("Не удалось подключиться к БД", zap.Error(err))
	}
	logger.Info("✓ Подключение к БД установлено")
	if err := store.RunMigrations("migrations"); err != nil {
		logger.Fatal("Не удалось применить миграции", zap.Error(err))
	}
	return store
}

func mustInitAccrualClient(logger *zap.Logger, baseURL string) *accrual.Client {
	client, err := accrual.NewClient(accrual.Config{
		BaseURL: baseURL,
		Logger:  logger,
	})
	if err != nil {
		logger.Fatal("Не удалось создать клиент системы начисления", zap.Error(err))
	}
	return client
}

func startAccrualWorker(logger *zap.Logger, store *storage.Storage, accrualSystemAddress string) *worker.AccrualWorker {
	accrualClient := mustInitAccrualClient(logger, accrualSystemAddress)
	accrualService := service.NewAccrualService(accrualClient, store, logger)
	w := worker.NewAccrualWorker(accrualService, logger, time.Second*2, 10)
	w.Start()
	return w
}

func startServer(store *storage.Storage, logger *zap.Logger, addr, JWTSecret string) *http.Server {
	authService := service.NewAuthService(JWTSecret, store)
	orderService := service.NewOrderService(store)
	balanceService := service.NewBalanceService(store)
	router := handler.NewRouter(JWTSecret, logger, authService, orderService, balanceService)

	server := &http.Server{
		Addr:         addr,
		Handler:      router.Setup(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Сервер слушает", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка сервера", zap.Error(err))
		}
	}()

	return server
}
