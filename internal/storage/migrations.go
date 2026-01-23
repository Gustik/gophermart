package storage

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations применяет все ожидающие миграции
func (s *Storage) RunMigrations(migrationsPath string) error {
	log.Println("Применение миграций...")

	// Создаём драйвер для БД
	driver, err := postgres.WithInstance(s.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать драйвер миграций: %w", err)
	}

	// Создаём экземпляр migrate
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать экземпляр migrate: %w", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil {
		// Если миграции уже применены, это не ошибка
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("✓ Миграции уже применены")
			return nil
		}
		return fmt.Errorf("не удалось применить миграции: %w", err)
	}

	log.Println("✓ Миграции успешно применены")
	return nil
}
