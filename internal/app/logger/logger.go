package logger

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	zap *zap.Logger
}

func NewZapLogger() *ZapLogger {
	zapLogger, _ := zap.NewProduction()
	return &ZapLogger{zap: zapLogger}
}

func (zl *ZapLogger) Debug(msg string, fields ...zap.Field) {
	zl.zap.Debug(msg, convertFields(fields)...)
}

func (zl *ZapLogger) Info(msg string, fields ...zap.Field) {
	zl.zap.Info(msg, convertFields(fields)...)
}

func (zl *ZapLogger) Warn(msg string, fields ...zap.Field) {
	zl.zap.Warn(msg, convertFields(fields)...)
}

func (zl *ZapLogger) Error(msg string, fields ...zap.Field) {
	zl.zap.Error(msg, convertFields(fields)...)
}

func convertFields(fields []zap.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.String)
	}
	return zapFields
}
