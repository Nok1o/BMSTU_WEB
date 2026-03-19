FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY ../go.mod go.sum ./
RUN go mod download
COPY .. .
COPY deploy/config /config
RUN CGO_ENABLED=0 GOOS=linux go build -o /smart-balancer ./cmd/main.go

# --- Финальная стадия ---
FROM alpine:latest
COPY --from=builder /smart-balancer /smart-balancer
COPY --from=builder /config /config
EXPOSE 9000
CMD ["/smart-balancer", "--config", "/config/config.yaml"]
