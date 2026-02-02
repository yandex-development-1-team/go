package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setEnvironment определяет и устанавливает конфигурацию Development/Production
func setEnvironment(env string) zap.Config {
	if env == "debug" {
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
func NewLogger(env string) *zap.Logger {
	config := setEnvironment(env)
	setLevel(&config, env)

	logger, err := config.Build()
	if err != nil {
		fallback, _ := zap.NewProduction()
		return fallback
	}

	return logger
}
