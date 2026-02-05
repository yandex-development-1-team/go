package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	PostgresURL      string `yaml:"postgres_url"`
	Port             int    `yaml:"port"`
	Environment      string `yaml:"environment"`
	PrometheusPort   int    `yaml:"prometheus_port"`
	LogLevel         string `yaml:"log_level"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("internal/config")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("port", 8080)
	viper.SetDefault("environment", "dev")
	viper.SetDefault("prometheus_port", 9090)
	viper.SetDefault("log_level", "info")

	// Set env vars mapping
	viper.AutomaticEnv()
	viper.BindEnv("telegram_bot_token", "BOT_TOKEN")
	viper.BindEnv("postgres_url", "POSTGRES_URL")
	viper.BindEnv("port", "PORT")
	viper.BindEnv("environment", "ENVIRONMENT")
	viper.BindEnv("prometheus_port", "PROMETHEUS_PORT")
	viper.BindEnv("log_level", "LOG_LEVEL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config.yml: %w", err)
		}
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error parsing config.yml: %w", err)
	}

	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func ValidateConfig(config *Config) error {

	if config.TelegramBotToken == "" {
		return fmt.Errorf("telegram_bot_token is empty")
	}

	if config.PostgresURL == "" {
		return fmt.Errorf("postgres_url is empty")
	}

	return nil

}
