# OpenAPI спецификация (YAML, разбивка по категориям)

- **Точка входа:** `openapi.yaml`
- **Общие ответы и ошибки:** `common.yaml` — подключается во все path-файлы через `$ref`
- **Пути по категориям:** `paths/auth.yaml`, `paths/boxes.yaml`, `paths/events.yaml`, и т.д.
- **Схемы по доменам:** `schemas/users.yaml`, `schemas/boxes.yaml`, и т.д.

## Правки по клиенту

- **Auth без Bearer:** `login`, `refresh`, `forgot-password`, `reset-password` — без токена (`security: []`).
- **logout:** добавлен `POST /api/v1/auth/logout` (с Bearer).
- **Восстановление пароля:** `POST /api/v1/auth/forgot-password` (email), `POST /api/v1/auth/reset-password` (token + new_password).
- **Пользователи:** `role`: `admin` | `manager`; `status`: `active` | `blocked` | `invited`.

Исходная спецификация в JSON: `../box.json`.
