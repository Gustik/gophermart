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
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Log configuration
	log.Printf("Starting gophermart server")
	log.Printf("Run address: %s", cfg.RunAddress)
	log.Printf("Database URI: %s", maskPassword(cfg.DatabaseURI))
	log.Printf("Accrual system: %s", cfg.AccrualSystemAddress)

	// TODO: Initialize database connection
	// TODO: Initialize router and handlers
	// TODO: Initialize worker

	// Создание сервера
	server := &http.Server{
		Addr: cfg.RunAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Gophermart API"))
		}),
	}

	go func() {
		log.Printf("Server listening on %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// maskPassword masks the password in database URI for logging
func maskPassword(uri string) string {
	// Simple masking for logging
	// TODO: implement proper masking
	return uri
}
