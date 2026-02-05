package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

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
func NewLogger(mode string, level string) {
	var err error
	config := setEnvironment(mode)
	setLevel(&config, level)

	logger, err = config.Build()
	if err != nil {
		logger, _ = zap.NewProduction()
	}
}

// Sync удаляет все буферизованные записи журнала
// Нужно вызывать Sync перед выходом
func Sync() error {
	return logger.Sync()
}

// Info логирование информационных сообщений
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error логирование ошибок
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Warn логирование предупреждений
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Debug логирование отладочной информации
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Fatal логирование критических ошибок
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Panic логирование ошибок с паникой
func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}
