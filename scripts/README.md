# Скрипты запуска бэкенда в Docker

Скрипты поднимают и останавливают бэкенд в Docker. Есть два режима: **api_only** (по умолчанию) и **bot**.

## Требования

- Docker и Docker Compose
- Для режима **bot** — файл `.env` в корне с переменной `BOT_TOKEN`

## Режимы

| Режим      | Команда           | Что поднимается                    |
|------------|-------------------|------------------------------------|
| **api_only** | `./scripts/run` или `./scripts/run api_only` | postgres, redis, api (миграции, health, метрики). Бот не запускается. |
| **bot**      | `./scripts/run bot` | postgres, redis, api и бот. Нужен `BOT_TOKEN` в `.env`. |

## Подготовка (только для режима bot)

```bash
cp .env.example .env
# Укажите BOT_TOKEN=ваш_токен_от_BotFather
```

Для **api_only** файл `.env` не обязателен.

## Запуск

Из корня проекта:

```bash
./scripts/run          # api_only: только API (postgres, redis, api)
./scripts/run bot      # с ботом (postgres, redis, api, bot)
```

Все сервисы в режиме хоста (`network_mode: host`), всё на **localhost**:

| Сервис   | Порт | Назначение                    |
|----------|------|-------------------------------|
| postgres | 5432 | БД                            |
| redis    | 6379 | Кэш/сессии                    |
| api      | 9090 | Health (`/health`), метрики (`/metrics`) — в обоих режимах |
| bot      | 9090 | Только в режиме `bot` (в этом режиме контейнер api не запускается) |

Логи бота в реальном времени:

```bash
docker compose logs -f bot
```

## Остановка

```bash
./scripts/stop
```

Выполняется `docker compose down`: останавливаются и удаляются контейнеры bot, redis, postgres. Том с данными PostgreSQL сохраняется.

## Переменные в .env

| Переменная       | Обязательна | Описание                          |
|------------------|-------------|-----------------------------------|
| `BOT_TOKEN`      | да          | Токен Telegram-бота               |
| `POSTGRES_USER`  | нет         | Пользователь БД (по умолчанию `bot_user`) |
| `POSTGRES_PASSWORD` | нет      | Пароль БД (по умолчанию `bot_password`)   |
| `POSTGRES_DB`    | нет         | Имя БД (по умолчанию `bot_db`)    |

В режиме **api_only** приложение запускается с `RUN_MODE=api_only`: выполняются миграции, поднимаются health и метрики на порту 9090, Telegram-бот не инициализируется. В режиме **bot** запускается полный процесс с ботом (BOT_TOKEN обязателен).

Конфиг: в Docker задаётся `CONFIG_FILE=/app/config/docker.yaml`; адреса `127.0.0.1` (localhost). Локально используется `config/config.yaml`.

**Примечание:** `network_mode: host` корректно работает на Linux. На macOS/Windows поведение может отличаться; при проблемах с подключением к БД/Redis проверьте документацию Docker Desktop.
