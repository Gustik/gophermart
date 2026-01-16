# Gophermart - Loyalty System

Накопительная система лояльности для интернет-магазина «Гофермарт».

## Требования

- Go 1.22+
- PostgreSQL 15+
- Docker & Docker Compose (опционально)

## Быстрый старт

### Локальный запуск

1. Запустить PostgreSQL:
```bash
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gophermart \
  -p 5432:5432 \
  postgres:15
```

2. Запустить систему начисления:
```bash
cd cmd/accrual
chmod +x accrual_linux_amd64
RUN_ADDRESS="localhost:8081" \
./accrual_linux_amd64
```

3. Запустить gophermart:
```bash
make run-flags
```

### Docker Compose
```bash
make docker-up
```

## Конфигурация

Сервис поддерживает конфигурирование через флаги и переменные окружения:

| Параметр | Флаг | Environment | По умолчанию |
|----------|------|-------------|--------------|
| Адрес сервера | `-a` | `RUN_ADDRESS` | `localhost:8080` |
| База данных | `-d` | `DATABASE_URI` | - |
| Система начисления | `-r` | `ACCRUAL_SYSTEM_ADDRESS` | `http://localhost:8081` |

## API Endpoints

- `POST /api/user/register` - регистрация пользователя
- `POST /api/user/login` - аутентификация пользователя
- `POST /api/user/orders` - загрузка номера заказа
- `GET /api/user/orders` - список заказов
- `GET /api/user/balance` - текущий баланс
- `POST /api/user/balance/withdraw` - списание баллов
- `GET /api/user/withdrawals` - история списаний

## Разработка

### Сборка
```bash
make build
```

### Тесты
```bash
make test
```

### Покрытие тестами
```bash
make test-coverage
```

### Форматирование
```bash
make fmt
```

## Структура проекта
```
gophermart/
├── cmd/gophermart/     # Entry point
├── internal/           # Private application code
│   ├── config/        # Configuration
│   ├── storage/       # Database layer
│   ├── handlers/      # HTTP handlers
│   ├── auth/          # Authentication
│   ├── models/        # Data models
│   └── accrual/       # Accrual client
├── migrations/        # Database migrations
```