// cmd/gophermart/main.go

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gustik/gophermart/internal/config"
	"github.com/Gustik/gophermart/internal/handler"
	"github.com/Gustik/gophermart/internal/service"
	"github.com/Gustik/gophermart/internal/storage"
	"github.com/Gustik/gophermart/internal/zaplog"
)

func main() {
	cfg := config.New()

	logger, err := zaplog.New("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Запуск сервера gophermart")
	logger.Sugar().Infof("Адрес сервера: %s", cfg.RunAddress)
	logger.Sugar().Infof("База данных: %s", maskPassword(cfg.DatabaseURI))
	logger.Sugar().Infof("Система начисления: %s", cfg.AccrualSystemAddress)

	// Подключение к БД
	store, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		logger.Sugar().Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer store.Close()

	logger.Info("✓ Подключение к БД установлено")

	if err := store.RunMigrations("migrations"); err != nil {
		logger.Sugar().Fatalf("Не удалось применить миграции: %v", err)
	}

	authService := service.NewAuthService(cfg.JWTSecret, store)
	router := handler.NewRouter(cfg.JWTSecret, logger, authService)

	// Создание HTTP сервера
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router.Setup(),
	}

	go func() {
		logger.Sugar().Infof("Сервер слушает на %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

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
	// Простая маскировка для примера
	// TODO: реализовать правильную маскировку
	return uri
}
