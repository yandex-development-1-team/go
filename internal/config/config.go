package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken  string         `mapstructure:"telegram_bot_token"`
	TelegramBotAPIUrl string         `mapstructure:"telegram_bot_api_url"`
	Telegram          Telegram       `mapstructure:"telegram"`
	AuthConfig        AuthConfig     `mapstructure:"auth_config"`
	DB                DatabaseConfig `mapstructure:"db"`
	Port              int            `mapstructure:"port"`
	Environment       string         `mapstructure:"environment"`
	PrometheusPort    int            `mapstructure:"prometheus_port"`
	LogLevel          string         `mapstructure:"log_level"`
	HostName          string         `mapstructure:"host_name"`
	Redis             RedisConfig    `mapstructure:"redis"`
	Session           SessionConfig  `mapstructure:"session"`
	MsgRPS            float64        `mapstructure:"msg_rps"`
	ApiRPS            float64        `mapstructure:"api_rps"`
	CacheSizeRPS      int            `mapstructure:"cache_size_rps"`
	APIOnly           bool           `mapstructure:"api_only"` // only API + metrics, no telegram bot
	CORS              CORSConfig     `mapstructure:"cors"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
	ApiUrl   string `mapstructure:"api_url"`
	Debug    bool   `mapstructure:"debug"`

	Proxy struct {
		Enabled bool   `mapstructure:"enabled"`
		Server  string `mapstructure:"server"`
		Port    string `mapstructure:"port"`
	} `mapstructure:"proxy"`
}

// CORSConfig — настройки CORS для HTTP API.
type CORSConfig struct {
	AllowOrigin      string `mapstructure:"allow_origin"`
	AllowMethods     string `mapstructure:"allow_methods"`
	AllowHeaders     string `mapstructure:"allow_headers"`
	ExposeHeaders    string `mapstructure:"expose_headers"`
	AllowCredentials string `mapstructure:"allow_credentials"`
	MaxAge           string `mapstructure:"max_age"`
}

type DatabaseConfig struct {
	PostgresURL string
	Name        string `mapstructure:"name"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	SslMode     string `mapstructure:"ssl_mode"`
	HostPort    string `mapstructure:"host_port"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	// Pool
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	// Timeouts
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

type AuthConfig struct {
	JWTSecret             string `mapstructure:"jwt_secret"`
	AccessTokenTTLMinutes int    `mapstructure:"access_token_ttl_minutes"`
	RefreshTokenTTLDays   int    `mapstructure:"refresh_token_ttl_days"`
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

	if config.Telegram.BotToken == "" && config.TelegramBotToken != "" {
		config.Telegram.BotToken = config.TelegramBotToken
	}
	if config.TelegramBotToken == "" && config.Telegram.BotToken != "" {
		config.TelegramBotToken = config.Telegram.BotToken
	}

	if config.DB.Name != "" {
		config.DB.PostgresURL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
			config.DB.User, config.DB.Password, config.DB.HostPort, config.DB.Name, config.DB.SslMode)
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("telegram.api_url", "https://api.telegram.org")
	v.SetDefault("port", 8080)
	v.SetDefault("environment", "dev")
	v.SetDefault("prometheus_port", 9090)
	v.SetDefault("log_level", "info")
	v.SetDefault("host_name", "unknown")

	v.SetDefault("db.ssl_mode", "disable")
	v.SetDefault("db.host_port", "localhost:5432")

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
	// BindEnv returns error only on invalid key; keys are fixed at compile time.
	_ = v.BindEnv("telegram_bot_token", "BOT_TOKEN")
	_ = v.BindEnv("postgres_url", "POSTGRES_URL")
	_ = v.BindEnv("port", "SERVER_PORT")

	_ = v.BindEnv("db.name", "POSTGRES_NAME")
	_ = v.BindEnv("db.user", "POSTGRES_USER")
	_ = v.BindEnv("db.password", "POSTGRES_PASSWORD")
	_ = v.BindEnv("db.ssl_mode", "DB_SSLMODE")
	_ = v.BindEnv("db.host_port", "DB_HOST_PORT")

	_ = v.BindEnv("redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")

	_ = v.BindEnv("prometheus_port", "PROMETHEUS_PORT")

	_ = v.BindEnv("environment", "ENVIRONMENT")

	_ = v.BindEnv("log_level", "LOG_LEVEL")
	_ = v.BindEnv("host_name", "HOSTNAME")
	_ = v.BindEnv("api_only", "API_ONLY")

	_ = v.BindEnv("auth_config.jwt_secret", "JWT_SECRET")
}

func validateConfig(config *Config) error {
	if !config.APIOnly && config.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}

	if config.DB.PostgresURL == "" {
		return fmt.Errorf("postgres_url is empty")
	}

	return nil
}
