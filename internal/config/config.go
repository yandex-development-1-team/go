package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken  string        `mapstructure:"telegram_bot_token"`
	TelegramBotAPIUrl string        `mapstructure:"telegram_bot_api_url"`
	PostgresURL       string        `mapstructure:"postgres_url"`
	Port              int           `mapstructure:"port"`
	Environment       string        `mapstructure:"environment"`
	PrometheusPort    int           `mapstructure:"prometheus_port"`
	LogLevel          string        `mapstructure:"log_level"`
	HostName          string        `mapstructure:"host_name"`
	Redis             RedisConfig   `mapstructure:"redis"`
	Session           SessionConfig `mapstructure:"session"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	// Пул
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	// Таймауты
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`

	// Retry
	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`
}

type SessionConfig struct {
	TTL time.Duration `mapstructure:"ttl"`
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
	setDefaults(v)

	// Set env vars mapping
	bindEnvs(v)

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

func setDefaults(v *viper.Viper) {

	v.SetDefault("telegram_bot_api_url", "https://api.telegram.org")
	v.SetDefault("port", 8080)
	v.SetDefault("environment", "dev")
	v.SetDefault("prometheus_port", 9090)
	v.SetDefault("log_level", "info")
	v.SetDefault("host_name", "unknown")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)

	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 2)
	v.SetDefault("redis.max_idle_conns", 5)
	v.SetDefault("redis.conn_max_idle_time", "5m")
	v.SetDefault("redis.conn_max_lifetime", "1h")

	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.min_retry_backoff", "8ms")
	v.SetDefault("redis.max_retry_backoff", "512ms")

	v.SetDefault("session.ttl", "24h")
}

func bindEnvs(v *viper.Viper) {

	v.BindEnv("telegram_bot_token", "BOT_TOKEN")
	v.BindEnv("postgres_url", "POSTGRES_URL")
	v.BindEnv("port", "PORT")
	v.BindEnv("environment", "ENVIRONMENT")
	v.BindEnv("prometheus_port", "PROMETHEUS_PORT")
	v.BindEnv("log_level", "LOG_LEVEL")
	v.BindEnv("host_name", "HOSTNAME")

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
