package internal

import (
	"fmt"
	"os"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Info(msg string, tags ...interface{})
	Debug(msg string, tags ...interface{})
	Warn(msg string, tags ...interface{})
	Error(msg string, tags ...interface{})
	Panic(msg string, tags ...interface{})
	Sync() error
	// WithFields returns a new logger with the given key-value pairs baked in.
	WithFields(tags ...interface{}) ILogger
}

type Logger struct {
	internalLogger *zap.Logger
}

func NewLogger(logLevel string) ILogger {
	var level zapcore.Level
	switch logLevel {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stderr, level)
	internalLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{internalLogger: internalLogger}
}

func (l *Logger) Info(msg string, tags ...interface{}) {
	l.internalLogger.Info(msg, tagsToZapFields(tags...)...)
}

func (l *Logger) Debug(msg string, tags ...interface{}) {
	l.internalLogger.Debug(msg, tagsToZapFields(tags...)...)
}

func (l *Logger) Warn(msg string, tags ...interface{}) {
	l.internalLogger.Warn(msg, tagsToZapFields(tags...)...)
}

func (l *Logger) Error(msg string, tags ...interface{}) {
	l.internalLogger.Error(msg, tagsToZapFields(tags...)...)
}

func (l *Logger) Panic(msg string, tags ...interface{}) {
	l.internalLogger.Panic(msg, tagsToZapFields(tags...)...)
}

func (l *Logger) Sync() error {
	return l.internalLogger.Sync()
}

func (l *Logger) WithFields(tags ...interface{}) ILogger {
	return &Logger{internalLogger: l.internalLogger.With(tagsToZapFields(tags...)...)}
}

func tagsToZapFields(tags ...interface{}) (field []zap.Field) {
	var zapFields []zap.Field

	for i := 0; i+1 < len(tags); i += 2 {
		key, didCast := tags[i].(string)
		value := tags[i+1]

		if didCast {
			zapFields = append(zapFields, zap.Any(key, value))
		} else {
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("invalid_key_%d", i), value))
		}
	}

	return zapFields
}
