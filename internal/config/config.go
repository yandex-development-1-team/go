package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Port        int
	Environment string
	LogLevel    string
}

// LoadConfig loads configuration from config file
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configPath)

	// Set defaults
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found - use defaults
			return parseConfig()
		}

		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return parseConfig()
}

func parseConfig() (*Config, error) {
	cfg := &Config{}

	// App
	cfg.App.Port = viper.GetInt("app.port")
	cfg.App.Environment = viper.GetString("app.environment")
	cfg.App.LogLevel = viper.GetString("app.log_level")

	return cfg, nil
}
