package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads variables from a .env file into the process environment.
// It is safe to call when .env does not exist (e.g. in production where
// real environment variables are injected by the platform).
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, relying on real environment variables")
	}
}

// getEnv returns the environment variable value or a fallback default.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
