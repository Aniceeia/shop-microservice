package logger

import (
	"context"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level Level
}

func NewLogger(level Level) *Logger {
	return &Logger{level: level}
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.level <= DEBUG {
		l.log(ctx, "DEBUG", msg, fields...)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.level <= INFO {
		l.log(ctx, "INFO", msg, fields...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.level <= WARN {
		l.log(ctx, "WARN", msg, fields...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...Field) {
	if l.level <= ERROR {
		l.log(ctx, "ERROR", msg, fields...)
	}
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	if l.level <= FATAL {
		l.log(ctx, "FATAL", msg, fields...)
		os.Exit(1)
	}
}

func (l *Logger) log(ctx context.Context, level, msg string, fields ...Field) {
	timestamp := time.Now().Format(time.RFC3339)
	requestID := getRequestID(ctx)

	log.Printf("[%s] %s %s %s %v", timestamp, level, requestID, msg, fields)
}

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return "unknown"
}

type Field struct {
	Key   string
	Value any
}

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}
