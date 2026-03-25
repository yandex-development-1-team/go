# Скрипты запуска

Обёртки над `make` (см. корневой `Makefile` и `.env.example`).

| Скрипт    | Действие                          |
|-----------|-----------------------------------|
| `./scripts/run`       | `make run-local` (полный стек в Docker, порты на 127.0.0.1) |
| `./scripts/run local-api` | `make run-local-api` (только Postgres и Redis) |
| `./scripts/stop`      | `make stop-local`                 |

Перед запуском: `cp .env.example .env` и заполните переменные.

Локальный бэкенд поверх инфраструктуры: `make run-local-api`, затем с `DB_HOST_PORT=127.0.0.1:5432` и `REDIS_ADDR=127.0.0.1:6379` (и при необходимости `API_ONLY=true`) выполните `make run`.
