package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string `mapstructure:"telegram_bot_token"`
	PostgresURL      string `mapstructure:"postgres_url"`
	Port             int    `mapstructure:"port"`
	Environment      string `mapstructure:"environment"`
	PrometheusPort   int    `mapstructure:"prometheus_port"`
	LogLevel         string `mapstructure:"log_level"`
}

func MustLoadConfig() *Config {
	v := viper.New()

	// set path to config file
	v.AddConfigPath("./internal/config")
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	// set defaults
	v.SetDefault("port", 8080)
	v.SetDefault("environment", "dev")
	v.SetDefault("prometheus_port", 9090)
	v.SetDefault("log_level", "info")

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	// env vars override file config
	v.AutomaticEnv()
	v.BindEnv("telegram_bot_token", "BOT_TOKEN")
	v.BindEnv("postgres_url", "POSTGRES_URL")
	v.BindEnv("port", "PORT")
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("prometheus_port", "PROMETHEUS_PORT")
	v.BindEnv("log_level", "LOG_LEVEL")

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		panic(err)
	}

	if err := ValidateConfig(config); err != nil {
		panic(err)
	}
	return config
}

func ValidateConfig(config *Config) error {
	if config.TelegramBotToken == "" {
		return errors.New("telegram_bot_token is empty")
	}
	if config.PostgresURL == "" {
		return errors.New("postgres_url is empty")
	}
	return nil
}
