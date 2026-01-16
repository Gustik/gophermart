// cmd/gophermart/main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gustik/gophermart/internal/config"
	"github.com/Gustik/gophermart/internal/storage"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	// Логирование конфигурации
	log.Printf("Запуск сервера gophermart")
	log.Printf("Адрес сервера: %s", cfg.RunAddress)
	log.Printf("База данных: %s", maskPassword(cfg.DatabaseURI))
	log.Printf("Система начисления: %s", cfg.AccrualSystemAddress)

	// Подключение к БД
	store, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer store.Close()

	log.Println("✓ Подключение к БД установлено")

	// Применение миграций
	if err := store.RunMigrations("migrations"); err != nil {
		log.Fatalf("Не удалось применить миграции: %v", err)
	}

	// TODO: Инициализировать роутер и handlers
	// TODO: Инициализировать worker

	// Создание HTTP сервера
	server := &http.Server{
		Addr: cfg.RunAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Gophermart API v0.2 - БД подключена"))
		}),
	}

	// Запуск сервера в горутине
	go func() {
		log.Printf("Сервер слушает на %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Остановка сервера...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Принудительная остановка сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}

// maskPassword маскирует пароль в URI для безопасного логирования
func maskPassword(uri string) string {
	// Простая маскировка для примера
	// TODO: реализовать правильную маскировку
	return uri
}
