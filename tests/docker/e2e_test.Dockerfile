FROM golang:1.24 AS tester

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
ENV ALLURE_OUTPUT_PATH="/app"

CMD ["go", "test", "-v", "./..."]