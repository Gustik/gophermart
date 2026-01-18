package config

import (
	"flag"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	// RunAddress - адрес и порт для запуска сервера
	RunAddress string

	// DatabaseURI - строка подключения к PostgreSQL
	DatabaseURI string

	// AccrualSystemAddress - URL системы начисления баллов
	AccrualSystemAddress string

	// JWTSecret - секретный ключ для подписи JWT токена
	JWTSecret string
}

// New создаёт новую конфигурацию из флагов и переменных окружения.
// Приоритет: флаги > переменные окружения > значения по умолчанию
func New() *Config {
	cfg := &Config{}

	// Определение флагов командной строки
	flag.StringVar(&cfg.RunAddress, "a", "", "Адрес и порт запуска сервера")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Строка подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес системы начисления")
	flag.StringVar(&cfg.JWTSecret, "j", "", "Секретный ключ для подписи JWT токена")
	flag.Parse()

	// Если флаг не установлен, берём значение из переменной окружения
	if cfg.RunAddress == "" {
		cfg.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}

	// Значения по умолчанию для локальной разработки
	if cfg.RunAddress == "" {
		cfg.RunAddress = "localhost:8070"
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = "http://localhost:8071"
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "very_strong_secret"
	}

	return cfg
}
