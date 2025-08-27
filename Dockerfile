# Multi-stage build для оптимизации размера образа
FROM golang:1.25-alpine AS builder

# Установка необходимых пакетов для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Создание рабочей директории
WORKDIR /build

# Копирование go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum* ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY *.go ./

# Сборка приложения с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(date +%Y%m%d-%H%M%S)" \
    -o stashapp-bot \
    .

# Финальный минимальный образ
FROM alpine:latest

# Метаданные образа
LABEL maintainer="StashApp Bot" \
      description="Telegram bot for StashApp" \
      version="1.0.0"

# Установка необходимых пакетов
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    ffmpeg \
    && rm -rf /var/cache/apk/*

# Создание непривилегированного пользователя
RUN addgroup -g 1000 -S botuser && \
    adduser -u 1000 -S botuser -G botuser

# Создание директорий для данных
RUN mkdir -p /app/DATA && \
    chown -R botuser:botuser /app

# Копирование бинарного файла из builder
COPY --from=builder /build/stashapp-bot /app/stashapp-bot

# Установка прав на исполнение
RUN chmod +x /app/stashapp-bot

# Переключение на непривилегированного пользователя
USER botuser

# Рабочая директория
WORKDIR /app

# Переменные окружения по умолчанию
ENV STASH_URL="" \
    STASH_API_KEY="" \
    TELEGRAM_BOT_TOKEN="" \
    TZ="Europe/Moscow" \
    DATA="/app/DATA"
# Точка монтирования для временных файлов
VOLUME ["/app/DATA"]

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep stashapp-bot || exit 1

# Запуск приложения
ENTRYPOINT ["/app/stashapp-bot"]