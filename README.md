# Backend Service

Сервис для многофункционального телеграм бота для аккаунтов из Яндекс Staff, предоставляющего бизнес-функции для внутреннних заказчиков и их внешних партнеров.
#### Поддерживаемый функционал:

1. Авторизация через Staff аккаунт Яндекса с использованием ролевой модели;
2. Преоставляет перечень коробочных решений по посещению мероприятий с возможностью выбора свободных дат и времени;
3. Скачивание гайда по посещению мероприятий;
4. Запрос спецпроекта;
5. Просмотр примеров спец.проектов;
6. Просмотр справочной информации о сервисе;
7. Связь с поддержкой.

---

## Содержание
- Введение
- Структура проекта
- Требования
- Конфигурация
- Быстрый старт
- Запуск (локально и Docker)
- Миграции базы данных
- Makefile цели
- Логи и метрики
- Разработка
- Тесты и линтинг
- Точки здоровья (health/endpoints)
- Траблшутинг

---

## Введение

Сервис написан на Go, включает HTTP API, работу с БД и миграциями, логирование, метрики и отдельный процесс бота. 
Запуск возможен локально, в Docker и через `docker-compose`.

---

## Структура проекта

```
├── main.go                 # Точка входа HTTP-сервиса (API)
├── cmd/
│   └── bot/
│       └── main.go         # Точка входа отдельного процесса бота
├── internal/
│   ├── bot/                # Логика бота: обработчики, клиенты, воркеры
│   ├── handlers/           # HTTP-обработчики (роуты, контроллеры)
│   ├── models/             # Доменные модели и DTO
│   ├── database/           # Подключение к БД, репозитории
│   ├── config/             # Чтение и валидация конфигурации
│   ├── logger/             # Инициализация и обертки логирования
│   └── metrics/            # Метрики
├── migrations/             # SQL/скрипты миграций базы данных
├── config/
│   └── config.yaml         # Основной конфигурационный файл
├── Dockerfile              # Сборка Docker-образа приложения
├── docker-compose.yml      # Сценарий запуска сервиса и зависимостей (БД и т.д.)
├── Makefile                # Сценарии автоматизации (build, run, test, lint, migrate)
└── go.mod / go.sum         # Модули и зависимости Go
```

---

## Требования

- Go 1.25+ (рекомендуется актуальная версия)
- Docker 24+ и Docker Compose (для контейнерного запуска)
- Доступ к СУБД (например, PostgreSQL)

---

## Конфигурация

Базовая конфигурация хранится в `config/config.yaml`. Значения могут переопределяться через переменные окружения.

Пример содержимого `config/config.yaml`:
```yaml
server:
  port: 8080
  readTimeout: 5s
  writeTimeout: 5s

database:
  maxOpenConns: 20
  maxIdleConns: 5
  connMaxLifetime: 30m

logger:
  level: info
  format: json

metrics:
  enabled: true
  endpoint: /metrics

bot:
  enabled: true
  token: "${BOT_TOKEN}"     # можно переопределить через окружение
```

Переменные окружения (примеры):
- `APP_ENV=local|dev|prod`
- `BOT_TOKEN=...`
- `DB_DSN=...` (доступ к БД, ссылка на запись в защищённом хранилище)
- `LOG_LEVEL=debug|info|warn|error`

Загрузчик конфигурации читает `config.yaml`, затем применяет ENV-опции.

---

## Быстрый старт

- Убедитесь, что корректно заполнен `config/config.yaml`
- Запустите сервис:
    - Локально: `go run ./main.go`
    - В Docker: `docker compose up --build`

---

## Запуск

### Локально (без Docker)
- Запуск API:
  ```bash
  go run ./main.go
  ```
- Запуск бота:
  ```bash
  go run ./cmd/bot/main.go
  ```

### Docker
- Собрать образ:
  ```bash
  docker build -t backend-service:latest .
  ```
- Запустить контейнер:
  ```bash
  docker run --rm -p 8080:8080 \
    -e APP_ENV=prod \
    -e BOT_TOKEN=... \
    backend-service:latest
  ```

### Docker Compose
- Запуск сервиса и зависимостей:
  ```bash
  docker compose up --build
  ```
- Остановка:
  ```bash
  docker compose down
  ```

Примечание: убедитесь, что в `docker-compose.yml` настроены сервисы (например, `db`) и корректные переменные окружения.

---

## Миграции базы данных

Для управления миграциями используется [goose](https://github.com/pressly/goose). Миграции автоматически применяются при запуске приложения.

- Создание миграции: 
  ```bash
  make migration-create NAME=название
  make migration-create NAME=create_users_table
  ```
- Откат последней миграции:
  ```bash
  make migration-rollback DB_DSN="postgres://user:pass@localhost:5432/db?sslmode=disable"
  ```

- Структура файла миграции:

-- +goose Up
CREATE TABLE ...;

-- +goose Down
DROP TABLE ...;

- Команды Makefile: 
    - make migration — справка
    - make migration-create NAME=<name> — создать миграцию
    - make migration-rollback DB_DSN=<dsn> — откатить последнюю

- Логи при запуске:
    - Current database version: 20240321120000
    - Found 2 pending migration(s)
    - Applied 2 migration(s)
    - Database version: 20240321120000 → 20240322123456

---

## Логи и метрики

- Логи:
    - Конфигурируются через `internal/logger` и `config.yaml`.
    - Уровни логирования: `debug`, `info`, `error`.

- Метрики:
    - Экспорт метрик Prometheus включается через `metrics.enabled`.
    - Эндпоинт метрик задаётся `metrics.endpoint` (например, `/metrics`).
    - В `internal/metrics` обычно находятся регистраторы, middleware и хэндлеры.

---

## Разработка

- Структура кода следует стандарту Go Modules.
- Бизнес‑логика и инфраструктурный код разделены по пакетам внутри `internal/`.
- Новые HTTP‑маршруты добавляйте в `internal/handlers`.
- Новые модели — в `internal/models`.
- Работу с БД (репозитории/ORM) — в `internal/database`.
- Конфиги и валидация — в `internal/config`.

Рекомендации:
- Соблюдайте контракты интерфейсов для удобства тестирования.
- Используйте контекст `context.Context` во всех внешних вызовах.
- Оборачивайте ошибки и логируйте с полями (structured logging).

---

## Тесты и линтинг

- Запуск тестов:
  ```bash
  go test ./... -race -cover
  ```
- Линтинг (пример с golangci-lint):
  ```bash
  golangci-lint run
  ```
- Форматирование:
  ```bash
  gofmt -w .
  go mod tidy
  ```

---

## Точки здоровья (health / readiness / liveness)

Рекомендуется:
- `/healthz` — liveness probe (проверка, что процесс жив).
- `/readyz` — readiness probe (проверка готовности зависимостей, например БД).
- `/metrics` — метрики (если включено).

Порты и пути настраиваются через `config.yaml`.

---

## Траблшутинг

- Проблема запуска:
    - Проверьте `config/config.yaml` и переменные окружения.
    - Убедитесь, что БД доступна и DSN корректен.
- Ошибки миграций:
    - Удостоверьтесь, что инструмент миграций установлен.
    - Сверьте права пользователя БД.
- Нет метрик:
    - Проверьте `metrics.enabled` и путь `metrics.endpoint`.
- Логи пустые или слишком многословные:
    - Настройте `logger.level` и `logger.format`.

---