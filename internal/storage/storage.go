// Package storage предоставляет функции для работы с базой данных
package storage

import (
	"context"
	"fmt"

	"github.com/Gustik/gophermart/internal/model"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const pgDuplicateErrorCode = "23505"

// Tx определяет интерфейс транзакции
type Tx interface {
	Rollback() error
	Commit() error
}

// Storage определяет методы для работы с хранилищем
type Storage interface {
	// Lifecycle methods
	Close() error
	RunMigrations(migrationsPath string) error
	DB() *sqlx.DB

	// User methods
	CreateUser(ctx context.Context, login, passwordHash string) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	// Order methods
	CreateOrder(ctx context.Context, userID int, number string) error
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	GetOrdersByUser(ctx context.Context, userID int) ([]model.Order, error)
	GetPendingOrders(ctx context.Context) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, tx Tx, orderNumber string, status model.OrderStatus, accrual *float32) error

	// Balance methods
	InitBalance(ctx context.Context, userID int) error
	AddBalance(ctx context.Context, tx Tx, userID int, amount *float32) error
	GetBalance(ctx context.Context, tx Tx, userID int) (*model.Balance, error)
	WithdrawBalance(ctx context.Context, tx Tx, userID int, sum float32) error
	CreateWithdrawal(ctx context.Context, tx Tx, userID int, orderNumber string, sum float32) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)

	// Transaction methods
	BeginTx(ctx context.Context) (Tx, error)
}

// storage представляет хранилище данных
type storage struct {
	db   *sqlx.DB
	psql squirrel.StatementBuilderType
}

// New создаёт новый экземпляр Storage и подключается к базе данных
func New(databaseURI string) (Storage, error) {
	db, err := sqlx.Connect("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("не удалось соединиться с БД: %w", err)
	}

	return &storage{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

// Close закрывает соединение с базой данных
func (s *storage) Close() error {
	return s.db.Close()
}

// DB возвращает объект базы данных
func (s *storage) DB() *sqlx.DB {
	return s.db
}

// BeginTx начинает транзакцию
func (s *storage) BeginTx(ctx context.Context) (Tx, error) {
	return s.db.BeginTxx(ctx, nil)
}

// getExecutor возвращает executor (транзакция или обычное соединение)
func (s *storage) getExecutor(tx Tx) sqlx.ExtContext {
	if tx != nil {
		if sqlxTx, ok := tx.(*sqlx.Tx); ok {
			return sqlxTx
		}
	}
	return s.db
}
