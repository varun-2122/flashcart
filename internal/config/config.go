package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	App      AppConfig
	Database DBConfig
	Cache    CacheConfig
	Logger   LoggerConfig
}

// AppConfig defines server settings.
type AppConfig struct {
	Env             string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
}

// DBConfig defines PostgreSQL pool settings.
type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

// CacheConfig defines Redis settings.
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// LoggerConfig defines Logging behavior.
type LoggerConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

// Load loads configuration from environment variables, falling back to defaults.
func Load() (*Config, error) {
	// Attempt to load .env file if available (ignore error if missing in production)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:             getEnv("APP_ENV", "development"),
			Port:            getEnv("APP_PORT", "8080"),
			ReadTimeout:     getDurationEnv("APP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("APP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getDurationEnv("APP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
			RequestTimeout:  getDurationEnv("APP_REQUEST_TIMEOUT", 5*time.Second),
		},
		Database: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "flashcart_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConns:        getInt32Env("DB_MAX_CONNS", 25),
			MinConns:        getInt32Env("DB_MIN_CONNS", 5),
			MaxConnIdleTime: getDurationEnv("DB_MAX_CONN_IDLE_TIME", 15*time.Minute),
			MaxConnLifetime: getDurationEnv("DB_MAX_CONN_LIFETIME", 1*time.Hour),
		},
		Cache: CacheConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
			PoolSize: getIntEnv("REDIS_POOL_SIZE", 10),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

// DSN builds PostgreSQL connection DSN string.
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// RedisAddr returns formatted host:port for Redis.
func (c *CacheConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func getInt32Env(key string, fallback int32) int32 {
	return int32(getIntEnv(key, int(fallback)))
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	d, err := time.ParseDuration(valStr)
	if err != nil {
		return fallback
	}
	return d
}
