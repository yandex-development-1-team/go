// Package logger предоставляет глобальный логгер приложения.
// Инициализация: вызвать NewLogger() при старте (например в main).
// Доступ: везде использовать logger.Info/Error/Warn/Debug или logger.L() для получения *zap.Logger.
package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global *zap.Logger
	mu     sync.RWMutex
)

// L возвращает глобальный логгер. До вызова NewLogger возвращает no-op логгер (записи отбрасываются).
// Использовать для передачи *zap.Logger в сторонний код или когда нужны методы zap напрямую.
func L() *zap.Logger {
	mu.RLock()
	l := global
	mu.RUnlock()
	if l != nil {
		return l
	}
	return zap.NewNop()
}

func setEnvironment(mode string) zap.Config {
	if mode == "dev" {
		return zap.NewDevelopmentConfig()
	}
	return zap.NewProductionConfig()
}

func setLevel(config *zap.Config, level string) {
	if level != "" {
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(level)); err == nil {
			config.Level = zap.NewAtomicLevelAt(l)
		}
	}
}

// NewLogger инициализирует глобальный логгер. Вызывать один раз при старте приложения (например в main).
func NewLogger(mode string, level string) {
	config := setEnvironment(mode)
	setLevel(&config, level)

	l, err := config.Build()
	if err != nil {
		l, _ = zap.NewProduction()
	}

	mu.Lock()
	global = l
	mu.Unlock()
}

// Sync сбрасывает буфер логгера. Вызывать перед выходом из приложения (например defer logger.Sync() в main).
func Sync() error {
	return L().Sync()
}

// Info — информационное сообщение.
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Error — ошибка.
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Warn — предупреждение.
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Debug — отладочное сообщение.
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Fatal — критическая ошибка (завершает процесс).
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Panic — логирование с паникой.
func Panic(msg string, fields ...zap.Field) {
	L().Panic(msg, fields...)
}
