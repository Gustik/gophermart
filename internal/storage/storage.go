// Package storage предоставляет функции для работы с базой данных
package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Storage представляет хранилище данных
type Storage struct {
	db *sql.DB
}

// New создаёт новый экземпляр Storage и подключается к базе данных
func New(databaseURI string) (*Storage, error) {
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение с БД: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	return &Storage{db: db}, nil
}

// Close закрывает соединение с базой данных
func (s *Storage) Close() error {
	return s.db.Close()
}

// DB возвращает объект базы данных
func (s *Storage) DB() *sql.DB {
	return s.db
}
