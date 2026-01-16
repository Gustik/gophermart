package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func New() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "", "Server run address (host:port)")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	if cfg.RunAddress == "" {
		cfg.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = "localhost:8080"
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = "http://localhost:8081"
	}

	return cfg, nil
}
