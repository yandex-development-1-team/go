# ============================================
# Builder Stage - компиляция Go приложения
# ============================================
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости (кешируется при неизменных go.mod/go.sum)
RUN go mod download

# Копируем исходный код
COPY internal/ ./internal/
COPY cmd/ ./cmd/
COPY migrations/ ./migrations/

# Компилируем приложение
# CGO_ENABLED=0 - статическая сборка без CGO
# -ldflags "-s -w" - удаляем отладочную информацию и таблицу символов (минификация)
# -o /app/bot - выходной файл
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w" \
    -o /app/bot \
    ./cmd/bot/main.go

# ============================================
# Runtime Stage - минимальный образ
# ============================================
FROM alpine:latest

# Устанавливаем CA сертификаты (для HTTPS запросов) и curl (для health check)
RUN apk --no-cache add ca-certificates curl

# Создаём non-root пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем скомпилированный бинарник из builder
COPY --from=builder /app/bot /app/bot

# Копируем миграции (могут понадобиться в runtime)
COPY --from=builder /build/migrations /app/migrations

# Устанавливаем права на выполнение
RUN chmod +x /app/bot

# Переключаемся на non-root пользователя
USER appuser

# Открываем порты
# Порт 9090: метрики (/metrics) и health check (/health) - используется в dev, feat/14, feat/17, feat/21
# Порт 8080: основной API (пока в разработке, закомментирован в коде)
EXPOSE 9090
# EXPOSE 8080

# Health check (проверка каждые 30 секунд)
# /health endpoint реализован в ветках: dev, feat/14, feat/17, feat/21
# После мержа соответствующей ветки, раскомментируйте:
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD curl -f http://localhost:9090/health || exit 1
# Примечание: порт зависит от конфигурации (обычно PrometheusPort: 9090)

# Entrypoint
ENTRYPOINT ["/app/bot"]
