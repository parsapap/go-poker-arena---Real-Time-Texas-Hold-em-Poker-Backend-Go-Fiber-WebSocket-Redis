// Package config centralizes loading and validation of all runtime
// configuration. Every value originates from environment variables so secrets
// never live in source. Load() fails fast (returns an error) when a critical
// value is missing or weak, allowing main to refuse to start.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// minJWTSecretLen is the minimum acceptable length for JWT_SECRET. Anything
// shorter is considered weak and rejected at startup.
const minJWTSecretLen = 16

// weakSecrets are obvious placeholder values that must never reach production.
var weakSecrets = map[string]struct{}{
	"secret":                                {},
	"changeme":                              {},
	"change-me":                             {},
	"your-super-secret-jwt-key-change-in-production": {},
	"test-secret":                           {},
	"test-secret-key":                       {},
}

// Config holds all runtime configuration for the server.
type Config struct {
	Env             string
	Port            string
	LogLevel        string
	JWTSecret       string
	AllowedOrigins  string
	AutoMigrate     bool
	BodyLimitBytes  int
	RequestTimeout  time.Duration
	APIRateLimit    int
	WSMaxConns      int

	DB    DBConfig
	Redis RedisConfig
}

// DBConfig holds PostgreSQL connection settings.
type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// IsProduction reports whether the server is running in a production-like env.
func (c *Config) IsProduction() bool {
	return c.Env != "development"
}

// Load reads configuration from the environment and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		Env:            getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		AutoMigrate:    getEnv("AUTO_MIGRATE", "false") == "true",
		BodyLimitBytes: getEnvInt("BODY_LIMIT_BYTES", 1<<20), // 1 MiB
		RequestTimeout: time.Duration(getEnvInt("REQUEST_TIMEOUT_SECONDS", 30)) * time.Second,
		APIRateLimit:   getEnvInt("API_RATE_LIMIT", 100),
		WSMaxConns:     getEnvInt("WS_MAX_CONNECTIONS", 5),
		DB: DBConfig{
			Host:         os.Getenv("POSTGRES_HOST"),
			Port:         getEnv("POSTGRES_PORT", "5432"),
			User:         os.Getenv("POSTGRES_USER"),
			Password:     os.Getenv("POSTGRES_PASSWORD"),
			Name:         os.Getenv("POSTGRES_DB"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLife:  time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 20),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate enforces that critical configuration is present and not weak.
func (c *Config) Validate() error {
	var problems []string

	// JWT secret: required everywhere; strength enforced in production.
	switch {
	case c.JWTSecret == "":
		problems = append(problems, "JWT_SECRET is required")
	case c.IsProduction() && len(c.JWTSecret) < minJWTSecretLen:
		problems = append(problems, fmt.Sprintf("JWT_SECRET must be at least %d characters in production", minJWTSecretLen))
	case c.IsProduction() && isWeakSecret(c.JWTSecret):
		problems = append(problems, "JWT_SECRET is a known weak/placeholder value; set a strong secret")
	}

	// Database credentials are required to connect.
	if c.DB.Host == "" {
		problems = append(problems, "POSTGRES_HOST is required")
	}
	if c.DB.User == "" {
		problems = append(problems, "POSTGRES_USER is required")
	}
	if c.DB.Name == "" {
		problems = append(problems, "POSTGRES_DB is required")
	}

	// Redis host is required.
	if c.Redis.Host == "" {
		problems = append(problems, "REDIS_HOST is required")
	}

	// In production, refuse a wildcard CORS origin.
	if c.IsProduction() && strings.TrimSpace(c.AllowedOrigins) == "*" {
		problems = append(problems, "ALLOWED_ORIGINS must not be '*' in production")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

// DSN builds the PostgreSQL connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, getEnv("POSTGRES_SSLMODE", "disable"),
	)
}

// Addr returns the Redis host:port address.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func isWeakSecret(s string) bool {
	_, ok := weakSecrets[strings.ToLower(strings.TrimSpace(s))]
	return ok
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// ErrMissingConfig is returned by helpers that need a value that is absent.
var ErrMissingConfig = errors.New("missing required configuration")
