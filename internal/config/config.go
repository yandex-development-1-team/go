package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken  string `mapstructure:"telegram_bot_token"`
	TelegramBotAPIUrl string `mapstructure:"telegram_bot_api_url"`

	DB                DatabaseConfig `mapstructure:"db"`
	Port              int            `mapstructure:"port"`
	Environment       string         `mapstructure:"environment"`
	PrometheusPort    int            `mapstructure:"prometheus_port"`
	LogLevel          string         `mapstructure:"log_level"`
	HostName          string         `mapstructure:"host_name"`
	Redis             RedisConfig    `mapstructure:"redis"`
	Session           SessionConfig  `mapstructure:"session"`
	MockClientEnabled bool           `mapstructure:"mock_client_enabled"`
	MockLocalDir      string         `mapstructure:"mock_local_dir"`
	MsgRPS            float64        `mapstructure:"msg_rps"`
	ApiRPS            float64        `mapstructure:"api_rps"`
	CacheSizeRPS      int            `mapstructure:"cache_size_rps"`
	GetUpdatesTimeout time.Duration  `mapstructure:"get_updates_timeout"`
	CORS              CORSConfig     `mapstructure:"cors"`
}

// CORSConfig — настройки CORS для HTTP API.
type CORSConfig struct {
	AllowOrigin      string `mapstructure:"allow_origin"`      // например "*" или "https://app.example.com"
	AllowMethods     string `mapstructure:"allow_methods"`     // "GET, POST, PUT, DELETE, OPTIONS, PATCH"
	AllowHeaders     string `mapstructure:"allow_headers"`     // заголовки запроса
	ExposeHeaders    string `mapstructure:"expose_headers"`    // заголовки, доступные клиенту в ответе
	AllowCredentials string `mapstructure:"allow_credentials"` // "true" / "false"
	MaxAge           string `mapstructure:"max_age"`           // кэш preflight в секундах, например "86400"
}

type DatabaseConfig struct {
	PostgresURL string
	Name        string `mapstructure:"name"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	SslMode     string `mapstructure:"ssl_mode"`
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

// DefaultConfigPaths — каталоги поиска конфига, если не задан CONFIG_FILE.
var DefaultConfigPaths = []string{"config", "."}

var (
	appCfg   Config
	loadOnce sync.Once
	loadErr  error
)

// GetConfig загружает конфиг. paths — каталоги или путь к файлу; если пусто, берётся CONFIG_FILE или DefaultConfigPaths.
func GetConfig(paths []string) (Config, error) {
	loadOnce.Do(func() {
		if p := os.Getenv("CONFIG_FILE"); p != "" {
			paths = []string{p}
		} else if len(paths) == 0 {
			paths = DefaultConfigPaths
		}
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
	v.SetConfigType("yaml")
	// Явный путь к файлу (например из CONFIG_FILE): один элемент, похож на файл
	if len(paths) == 1 && (strings.Contains(paths[0], ".yaml") || strings.Contains(paths[0], ".yml")) {
		v.SetConfigFile(paths[0])
	} else {
		v.SetConfigName("config")
		if len(paths) > 0 {
			for _, path := range paths {
				v.AddConfigPath(path)
			}
		} else {
			v.AddConfigPath("config")
			v.AddConfigPath(".")
		}
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

	if u := v.GetString("postgres_url"); u != "" {
		config.DB.PostgresURL = u
	} else if config.DB.Name != "" {
		config.DB.PostgresURL = fmt.Sprintf("postgres://%s:%s@db:5432/%s?sslmode=%s",
			config.DB.User, config.DB.Password, config.DB.Name, config.DB.SslMode)
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

	v.SetDefault("db.ssl_mode", "disable")

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
	v.SetDefault("get_updates_timeout", "30s")

	v.SetDefault("cors.allow_origin", "*")
	v.SetDefault("cors.allow_methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	v.SetDefault("cors.allow_headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	v.SetDefault("cors.expose_headers", "Content-Length, Content-Type, Date, X-Total-Count")
	v.SetDefault("cors.allow_credentials", "true")
	v.SetDefault("cors.max_age", "86400")
}

func bindEnvs(v *viper.Viper) {
	v.BindEnv("telegram_bot_token", "BOT_TOKEN")
	v.BindEnv("postgres_url", "POSTGRES_URL")
	v.BindEnv("port", "SERVER_PORT")

	viper.BindEnv("db.name", "DB_NAME")
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.ssl_mode", "DB_SSLMODE")

	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	v.BindEnv("prometheus_port", "PROMETHEUS_PORT")

	v.BindEnv("environment", "ENVIRONMENT")

	v.BindEnv("log_level", "LOG_LEVEL")
	v.BindEnv("host_name", "HOSTNAME")
	v.BindEnv("mock_client_enabled", "MOCK_CLIENT_ENABLED")
	v.BindEnv("mock_local_dir", "MOCK_LOCAL_DIR")
	v.BindEnv("get_updates_timeout", "GET_UPDATES_TIMEOUT")

	v.BindEnv("cors.allow_origin", "CORS_ALLOW_ORIGIN")
	v.BindEnv("cors.allow_methods", "CORS_ALLOW_METHODS")
	v.BindEnv("cors.allow_headers", "CORS_ALLOW_HEADERS")
	v.BindEnv("cors.expose_headers", "CORS_EXPOSE_HEADERS")
	v.BindEnv("cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")
	v.BindEnv("cors.max_age", "CORS_MAX_AGE")
}

func validateConfig(config *Config) error {
	if os.Getenv("RUN_MODE") != "api_only" && config.TelegramBotToken == "" {
		return fmt.Errorf("telegram_bot_token is empty")
	}
	if config.DB.PostgresURL == "" {
		return fmt.Errorf("postgres_url is empty")
	}

	return nil
}
