# Стейдж сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

COPY file_service/deploy/scripts/setup-minio-buckets.sh /usr/local/bin/setup-minio-buckets.sh

# Копируем весь проект в билд-контейнер
COPY . .

# Переходим в папку file_service
WORKDIR /app/file_service

# Собираем сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o file_service ./internal/server.go

# Стейдж финальный
FROM debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y \
    curl \
    bash \
    && curl -sL https://dl.min.io/client/mc/release/linux-amd64/mc -o /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc \
    && apt-get clean

COPY --from=builder /app/file_service/file_service .

COPY --from=builder /app/file_service/deploy/config ./config

COPY --from=builder /usr/local/bin/setup-minio-buckets.sh /usr/local/bin/setup-minio-buckets.sh

# 12. Делаем скрипты исполнимыми
RUN chmod +x /usr/local/bin/setup-minio-buckets.sh

# Добавь перед CMD для вывода путей
RUN ls -l /usr/local/bin/setup-minio-buckets.sh

RUN echo "SCHEME=$SCHEME" && \
    if [ "$SCHEME" == "https" ]; then \
        printf "\nMINIO_SCHEME=https" >> /etc/environment && \
        printf "\nMINIO_PUBLIC_ENDPOINT=$DOMAIN/minio" >> /etc/environment; \
    else \
        printf "\nMINIO_SCHEME=http" >> /etc/environment && \
        printf "\nMINIO_PUBLIC_ENDPOINT=localhost:9000" >> /etc/environment; \
    fi

# Загружаем переменные в окружение контейнера
ARG SCHEME
ARG MINIO_PUBLIC_ENDPOINT

ENV MINIO_SCHEME=${SCHEME}
ENV MINIO_PUBLIC_ENDPOINT=${MINIO_PUBLIC_ENDPOINT}

EXPOSE 8081

# 13. Указываем команду для запуска
CMD ["/bin/bash", "-c", "/usr/local/bin/setup-minio-buckets.sh && ./file_service"]

#ENTRYPOINT ["./file_service"]
