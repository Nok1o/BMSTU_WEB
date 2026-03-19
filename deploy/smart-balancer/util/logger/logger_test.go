//go:build unit

package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLoggerOutput(t *testing.T) {
	type logFunc func(ctx context.Context, format string, args ...interface{})

	testCases := []struct {
		name          string
		level         string
		ctx           context.Context
		logFunc       logFunc
		msg           string
		args          []interface{}
		expectedParts []string
	}{
		{
			name:    "Info with Request ID",
			level:   "info",
			ctx:     ContextWithRequestID(context.Background(), "test-req-123"),
			logFunc: Infof,
			msg:     "simple info message",
			args:    nil,
			expectedParts: []string{
				"[INFO]",
				"(logger_test.go:",
				"[test-req-123]",
				"simple info message",
			},
		},
		{
			name:    "Warn without Request ID",
			level:   "warn",
			ctx:     context.Background(),
			logFunc: Warnf,
			msg:     "a warning occurred",
			args:    nil,
			expectedParts: []string{
				"[WARNING]",
				"(logger_test.go:",
				"[unknown request-id]",
				"a warning occurred",
			},
		},
		{
			name:    "Error with formatted message",
			level:   "error",
			ctx:     ContextWithRequestID(context.Background(), "err-req-456"),
			logFunc: Errorf,
			msg:     "error processing payment %d",
			args:    []interface{}{12345},
			expectedParts: []string{
				"[ERROR]",
				"(logger_test.go:",
				"[err-req-456]",
				"error processing payment 12345",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			InitLogger(tc.level, &output)

			tc.logFunc(tc.ctx, tc.msg, tc.args...)

			result := output.String()

			for _, part := range tc.expectedParts {
				if !strings.Contains(result, part) {
					t.Errorf("Log output for '%s' did not contain expected part '%s'.\nGot: %s", tc.name, part, result)
				}
			}
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	ctx := context.Background()

	t.Run("DEBUG message should be ignored on INFO level", func(t *testing.T) {
		var output bytes.Buffer
		InitLogger("info", &output)

		Debugf(ctx, "this is a debug message")

		if output.Len() > 0 {
			t.Errorf("Expected no output for DEBUG message on INFO level, but got: %s", output.String())
		}
	})

	t.Run("INFO message should be logged on INFO level", func(t *testing.T) {
		var output bytes.Buffer
		InitLogger("info", &output)

		Infof(ctx, "this is an info message")

		if output.Len() == 0 {
			t.Errorf("Expected output for INFO message on INFO level, but got none")
		}
	})

	t.Run("WARN message should be logged on INFO level", func(t *testing.T) {
		var output bytes.Buffer
		InitLogger("info", &output)

		Warnf(ctx, "this is a warning")

		if output.Len() == 0 {
			t.Errorf("Expected output for WARN message on INFO level, but got none")
		}
	})

	t.Run("INFO message should be ignored on WARN level", func(t *testing.T) {
		var output bytes.Buffer
		InitLogger("warn", &output)

		Infof(ctx, "this is an info message")

		if output.Len() > 0 {
			t.Errorf("Expected no output for INFO message on WARN level, but got: %s", output.String())
		}
	})
}

func TestContextWithRequestID(t *testing.T) {
	testID := "my-unique-id"
	ctx := context.Background()

	ctxWithID := ContextWithRequestID(ctx, testID)

	val, ok := ctxWithID.Value(requestIDKey).(string)
	if !ok {
		t.Fatal("Value from context is not a string or not found")
	}

	if val != testID {
		t.Errorf("Expected request ID to be '%s', but got '%s'", testID, val)
	}
}
