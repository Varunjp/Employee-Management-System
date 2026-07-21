package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret     string
	JWTExpiration time.Duration

	AdminUsername string
	AdminPassword string
}

// Load reads a .env file if present (local development convenience) and
// then builds a Config from environment variables, applying sane defaults
// wherever a variable is not set. In containers, real environment
// variables (e.g. from docker-compose) always take precedence over
// anything in a .env file.
func Load() (*Config, error) {
	_ = godotenv.Load() // ignored: absence of .env is expected in production

	jwtExpMinutes, err := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "60"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_EXPIRATION_MINUTES: %w", err)
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid REDIS_DB: %w", err)
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "employee_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiration: time.Duration(jwtExpMinutes) * time.Hour,

		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET must be set")
	}

	return cfg, nil
}

// DSN builds a libpq-style Postgres connection string from the config.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
