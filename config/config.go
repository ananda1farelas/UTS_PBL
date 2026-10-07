package config

import (
	"fmt"
	"time"
)

// Config holds all application configuration loaded from environment
// variables / .env file.
type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret    string
	JWTExpiresIn time.Duration

	LoginRateLimitMax    int
	LoginRateLimitWindow time.Duration
}

var cfg *Config

// Load reads environment variables (after LoadEnv has populated them from
// .env) and builds the global Config instance. Call it once at startup.
func Load() *Config {
	LoadEnv()

	cfg = &Config{
		AppPort: getEnv("APP_PORT", "3000"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "siakad_mini"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:    getEnv("JWT_SECRET", "secret"),
		JWTExpiresIn: parseDuration(getEnv("JWT_EXPIRES_IN", "24h"), 24*time.Hour),

		LoginRateLimitMax:    parseInt(getEnv("LOGIN_RATE_LIMIT_MAX", "5"), 5),
		LoginRateLimitWindow: parseDuration(getEnv("LOGIN_RATE_LIMIT_WINDOW", "1m"), time.Minute),
	}

	return cfg
}

// Get returns the already-loaded config. Panics if Load() has not run yet,
// since that would mean the app started in an invalid state.
func Get() *Config {
	if cfg == nil {
		panic("config: Get() called before Load()")
	}
	return cfg
}

// DSN builds the PostgreSQL connection string used by database/sql + lib/pq.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func parseInt(value string, fallback int) int {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
