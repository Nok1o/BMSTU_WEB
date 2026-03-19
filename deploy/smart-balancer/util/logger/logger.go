// file: logger/logger.go
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type CustomFormatter struct{}
type ctxKey string

const requestIDKey ctxKey = "requestID"

var loggerInitialized = false

// Format форматирует запись лога.
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 37 // Белый
	case logrus.WarnLevel:
		levelColor = 33 // Желтый
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // Красный
	default:
		levelColor = 36 // Голубой
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("\x1b[90m%s\x1b[0m ", entry.Time.Format(time.TimeOnly)))

	if caller, ok := entry.Data["caller"]; ok {
		b.WriteString(fmt.Sprintf(" \x1b[37m(%s)\x1b[0m ", caller))
	}

	if requestID, ok := entry.Data["request_id"]; ok {
		b.WriteString(fmt.Sprintf("\x1b[35m[%s]\x1b[0m ", requestID))
	} else {
		b.WriteString(fmt.Sprintf("\x1b[35m[%s]\x1b[0m ", "unknown request-id"))
	}

	levelText := strings.ToUpper(entry.Level.String())
	b.WriteString(fmt.Sprintf("\x1b[%dm[%s]\x1b[0m", levelColor, levelText))

	b.WriteString(fmt.Sprintf(" %s", entry.Message))

	b.WriteByte('\n')

	return []byte(b.String()), nil
}

func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// InitLogger инициализирует глобальный логгер logrus.
func InitLogger(levelStr string, out io.Writer) {
	loggerInitialized = true
	logrus.SetOutput(out)
	logrus.SetFormatter(&CustomFormatter{})

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
		logrus.Warnf("Некорректный уровень логирования: '%s'. Используется 'info'.", levelStr)
	}
	logrus.SetLevel(level)
}

// getEntry - создает запись лога, обогащенную полями из контекста и информацией о вызывающем.
func getEntry(ctx context.Context) *logrus.Entry {
	entry := logrus.NewEntry(logrus.StandardLogger())

	// Пропускаем 4 фрейма: Callers, getCaller, getEntry и саму функцию логирования (e.g., Infof)
	if caller := getCaller(4); caller != nil {
		entry = entry.WithField("caller", fmt.Sprintf("%s:%d", path.Base(caller.File), caller.Line))
	}

	if id, ok := ctx.Value(requestIDKey).(string); ok {
		entry = entry.WithField("request_id", id)
	}

	return entry
}

// getCaller находит правильное место вызова в стеке.
func getCaller(skip int) *runtime.Frame {
	pcs := make([]uintptr, 25)
	depth := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	pathSep := string(filepath.Separator)

	for {
		f, more := frames.Next()
		if !strings.Contains(f.File, "sirupsen"+pathSep+"logrus") &&
			!strings.Contains(f.File, "logger"+pathSep+"logger.go") {
			return &f
		}
		if !more {
			break
		}
	}
	return nil
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	if !loggerInitialized {
		log.Fatalf("Logger was not initialized")
	}
	getEntry(ctx).Debugf(format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	if !loggerInitialized {
		log.Fatalf("Logger was not initialized")
	}
	getEntry(ctx).Infof(format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	if !loggerInitialized {
		log.Fatalf("Logger was not initialized")
	}
	getEntry(ctx).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	if !loggerInitialized {
		log.Fatalf("Logger was not initialized")
	}
	getEntry(ctx).Errorf(format, args...)
}
