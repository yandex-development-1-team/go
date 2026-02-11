package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string `mapstructure:"telegram_bot_token"`
	PostgresURL      string `mapstructure:"postgres_url"`
	Port             int    `mapstructure:"port"`
	Environment      string `mapstructure:"environment"`
	PrometheusPort   int    `mapstructure:"prometheus_port"`
	LogLevel         string `mapstructure:"log_level"`
	HostName         string `mapstructure:"host_name"`
}

var (
	appCfg   Config
	loadOnce sync.Once
	loadErr  error
)

func GetConfig(paths []string) (Config, error) {
	loadOnce.Do(func() {
		cfg, err := loadConfig(paths)
		if err != nil {
			loadErr = err
			return
		}
		appCfg = *cfg
	})
	return appCfg, loadErr
}

func loadConfig(paths []string) (*Config, error) {

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if len(paths) > 0 {
		for _, path := range paths {
			v.AddConfigPath(path)
		}
	} else {
		v.AddConfigPath("config")
		v.AddConfigPath(".")
	}

	// Set defaults
	v.SetDefault("port", 8080)
	v.SetDefault("environment", "dev")
	v.SetDefault("prometheus_port", 9090)
	v.SetDefault("log_level", "info")
	v.SetDefault("host_name", "unknown")

	// Set env vars mapping
	v.BindEnv("telegram_bot_token", "BOT_TOKEN")
	v.BindEnv("postgres_url", "POSTGRES_URL")
	v.BindEnv("port", "PORT")
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("prometheus_port", "PROMETHEUS_PORT")
	v.BindEnv("log_level", "LOG_LEVEL")
	v.BindEnv("host_name", "HOSTNAME")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found")
		}
		return nil, fmt.Errorf("error reading config.yml: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error parsing config.yml: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config *Config) error {
	if config.TelegramBotToken == "" {
		return fmt.Errorf("telegram_bot_token is empty")
	}

	if config.PostgresURL == "" {
		return fmt.Errorf("postgres_url is empty")
	}

	return nil
}
