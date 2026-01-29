// Package storage предоставляет функции для работы с базой данных
package storage

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const pgDuplicateErrorCode = "23505"

// Storage представляет хранилище данных
type Storage struct {
	db   *sqlx.DB
	psql squirrel.StatementBuilderType
}

// New создаёт новый экземпляр Storage и подключается к базе данных
func New(databaseURI string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("не удалось соединиться с БД: %w", err)
	}

	return &Storage{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

// Close закрывает соединение с базой данных
func (s *Storage) Close() error {
	return s.db.Close()
}

// DB возвращает объект базы данных
func (s *Storage) DB() *sqlx.DB {
	return s.db
}
