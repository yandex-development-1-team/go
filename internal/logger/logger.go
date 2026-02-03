package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
func NewLogger(mode string, level string) *zap.Logger {
	config := setEnvironment(mode)
	setLevel(&config, level)

	logger, err := config.Build()
	if err != nil {
		fallback, _ := zap.NewProduction()
		return fallback
	}

	return logger
}
