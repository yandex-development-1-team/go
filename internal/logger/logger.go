package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger структура логгера
type Logger struct {
	Logger *zap.Logger
}

// setEnvironment определяет и устанавливает конфигурацию Development/Production
func setEnvironment(mode string) zap.Config {
	if mode == "dev" {
		return zap.NewDevelopmentConfig()
	}

	return zap.NewProductionConfig()
}

// setLevel устанавливает уровень логгирования
func setLevel(config *zap.Config, env string) {
	if env != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(env)); err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}
}

// NewLogger создает и настраивает логгер
func NewLogger(mode string, level string) *Logger {
	config := setEnvironment(mode)
	setLevel(&config, level)

	logger, err := config.Build()
	if err != nil {
		fallback, _ := zap.NewProduction()
		return &Logger{
			Logger: fallback,
		}
	}

	return &Logger{
		Logger: logger,
	}
}

// Info логирование информационных сообщений
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Error логирование ошибок
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Warn логирование предупреждений
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Debug логирование отладочной информации
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Fatal логирование критических ошибок
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}
