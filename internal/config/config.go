package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	RunAddress           string // адрес и порт для запуска сервера
	DatabaseURI          string // строка подключения к PostgreSQL
	AccrualSystemAddress string // URL системы начисления баллов
	JWTSecret            string // секретный ключ для подписи JWT токена
}

// New создаёт новую конфигурацию из флагов и переменных окружения.
// Приоритет: флаги > переменные окружения > значения по умолчанию
func New() *Config {
	cfg := &Config{}

	// Определение флагов командной строки, если значения нет то с переменной окружения
	flag.StringVar(&cfg.RunAddress, "a", getEnv("RUN_ADDRESS", "localhost:8070"), "Адрес и порт запуска сервера")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnv("DATABASE_URI", "postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable"), "Строка подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", getEnv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8071"), "Адрес системы начисления")
	flag.StringVar(&cfg.JWTSecret, "j", getEnv("JWT_SECRET", "very_strong_secret"), "Секретный ключ для JWT")
	flag.Parse()

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Info возвращает информацию о конфигурации для логирования
func (c *Config) Info() string {
	return fmt.Sprintf(
		"Конфигурация сервера:\n"+
			"  Адрес сервера: %s\n"+
			"  База данных: %s\n"+
			"  Система начисления: %s",
		c.RunAddress,
		maskPassword(c.DatabaseURI),
		c.AccrualSystemAddress,
	)
}

// maskPassword маскирует пароль в URI
func maskPassword(uri string) string {
	if uri == "" {
		return ""
	}

	// Парсим URI
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	// Если есть пароль, заменяем на ***
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		parsedURL.User = url.UserPassword(username, "***")
	}

	return parsedURL.String()
}
