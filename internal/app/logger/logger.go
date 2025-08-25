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

// func (zl *ZapLogger) Debug(msg string, fields ...Field) {
// 	zl.zap.Debug(msg, convertFields(fields)...)
// }

// func (zl *ZapLogger) Info(msg string, fields ...Field) {
// 	zl.zap.Info(msg, convertFields(fields)...)
// }

// func (zl *ZapLogger) Warn(msg string, fields ...Field) {
// 	zl.zap.Warn(msg, convertFields(fields)...)
// }

// func (zl *ZapLogger) Error(msg string, fields ...Field) {
// 	zl.zap.Error(msg, convertFields(fields)...)
// }

// func convertFields(fields []Field) []zap.Field {
// 	zapFields := make([]zap.Field, len(fields))
// 	for i, f := range fields {
// 		zapFields[i] = zap.Any(f.Key, f.Value)
// 	}
// 	return zapFields
// }
