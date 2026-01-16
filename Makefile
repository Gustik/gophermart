.PHONY: build run test clean docker-up docker-down migrate-up migrate-down

# Параметры Go
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMIGRATECMD=~/go/bin/migrate
BINARY_NAME=gophermart
BINARY_PATH=cmd/gophermart

# Параметры БД
DB_URI=postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable

# Сборка проекта
build:
	cd $(BINARY_PATH) && $(GOBUILD) -o $(BINARY_NAME) -v

# Запуск локально
run: build
	cd $(BINARY_PATH) && ./$(BINARY_NAME)

# Запуск с флагами
run-flags:
	$(GOCMD) run $(BINARY_PATH)/main.go \
		-a "localhost:8080" \
		-d "$(DB_URI)" \
		-r "http://localhost:8081"

# Запуск тестов
test:
	$(GOTEST) -v -cover ./...

# Покрытие тестами
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Отчёт о покрытии: coverage.html"

# Очистка
clean:
	rm -f $(BINARY_PATH)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

# Обновление зависимостей
tidy:
	go mod tidy

# Docker Compose
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

# Миграции (через golang-migrate CLI, если установлен)
migrate-up:
	$(GOMIGRATECMD) -path migrations -database "$(DB_URI)" up

migrate-down:
	$(GOMIGRATECMD) -path migrations -database "$(DB_URI)" down

# Создать новую миграцию
migrate-create:
	@read -p "Название миграции: " name; \
	$(GOMIGRATECMD) create -ext sql -dir migrations -seq $$name

# Форматирование кода
fmt:
	$(GOCMD) fmt ./...

# Справка
help:
	@echo "Доступные команды:"
	@echo "  build          - Собрать бинарник"
	@echo "  run            - Собрать и запустить"
	@echo "  run-flags      - Запустить с примерами флагов"
	@echo "  test           - Запустить тесты"
	@echo "  test-coverage  - Тесты с покрытием"
	@echo "  clean          - Очистить артефакты сборки"
	@echo "  docker-up      - Запустить через docker-compose"
	@echo "  docker-down    - Остановить docker-compose"
	@echo "  migrate-up     - Применить миграции"
	@echo "  migrate-down   - Откатить миграции"
	@echo "  migrate-create - Создать новую миграцию"
	@echo "  fmt            - Форматировать код"