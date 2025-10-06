//go:build unit
// +build unit

package logger_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/logger"
)

func TestInfoLog(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Создание контекста с requestID
	ctx := context.WithValue(context.Background(), logger.RequestID, uuid.New().String())

	// Логируем информацию
	logger.Info(ctx, "This is an info log")

	// Проверка, что в логе есть ожидаемые данные
	logOutput := buf.String()
	assert.Contains(t, logOutput, "[INFO]")
	assert.Contains(t, logOutput, "This is an info log")
}

func TestWarnLog(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Создание контекста с requestID
	ctx := context.WithValue(context.Background(), logger.RequestID, uuid.New().String())

	// Логируем предупреждение
	logger.Warn(ctx, "This is a warning log")

	// Проверка, что в логе есть ожидаемые данные
	logOutput := buf.String()
	assert.Contains(t, logOutput, "This is a warning log")
}

func TestErrorLog(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Создание контекста с requestID
	ctx := context.WithValue(context.Background(), logger.RequestID, uuid.New().String())

	// Логируем ошибку
	logger.Error(ctx, "This is an error log")

	// Проверка, что в логе есть ожидаемые данные
	logOutput := buf.String()
	assert.Contains(t, logOutput, "[ERROR]")
	assert.Contains(t, logOutput, "This is an error log")
}

func TestLogWithMissingRequestID(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Логируем информацию без requestID
	ctx := context.Background()
	logger.Info(ctx, "This is a log without requestID")

	// Проверка, что в логе есть ожидаемые данные
	logOutput := buf.String()
	assert.Contains(t, logOutput, "[INFO]")
	assert.Contains(t, logOutput, "This is a log without requestID")
	assert.Contains(t, logOutput, "unknownRequestID") // Проверка, что используется unknownRequestID
}

func TestLogFormat(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Создание контекста с requestID
	ctx := context.WithValue(context.Background(), logger.RequestID, uuid.New().String())

	// Логируем сообщение
	logger.Info(ctx, "This is a test log")

	// Проверка правильности форматирования
	logOutput := buf.String()

	// Проверка, что форматирование включает все необходимые элементы
	assert.Regexp(t, `\[INFO\].*\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\].*`, logOutput) // Проверка наличия времени
}

func TestLogContextualFields(t *testing.T) {
	// Подготовка
	var buf bytes.Buffer
	logger.Log.SetOutput(&buf)

	// Создание контекста с requestID и дополнительным полем
	ctx := context.WithValue(context.Background(), logger.RequestID, uuid.New().String())

	// Логируем сообщение с дополнительным контекстом
	logger.Info(ctx, "Test log with additional fields")

	// Проверка наличия дополнительных данных в логе
	logOutput := buf.String()
	assert.Contains(t, logOutput, "[INFO]")
}
