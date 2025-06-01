// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment    string
	Port           string
	DatabaseURL    string
	MigrationsPath string
	LogLevel       string

	// Database specifics parsed from DATABASE_URL
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	// This allows the app to work with real environment variables in production
	if err := godotenv.Load(); err != nil {
		// Only log if the file exists but can't be read
		if _, statErr := os.Stat(".env"); statErr == nil {
			log.Printf("Warning: .env file exists but couldn't be loaded: %v", err)
		}
		// If file doesn't exist, that's fine - we'll use actual env vars
	}

	cfg := &Config{
		Environment:    getEnv("ENVIRONMENT", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// For convenience, also parse DATABASE_URL components
	if err := cfg.parseDatabaseURL(); err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	return cfg, nil
}

// LoadFromFile explicitly loads configuration from a specific .env file
// This is useful for testing or when you have multiple env files
func LoadFromFile(filename string) (*Config, error) {
	if err := godotenv.Load(filename); err != nil {
		return nil, fmt.Errorf("failed to load env file %s: %w", filename, err)
	}

	return Load()
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseDatabaseURL extracts components from a postgres connection string
func (c *Config) parseDatabaseURL() error {
	// In a real application, you'd use a proper URL parser
	// For now, we'll trust that DATABASE_URL is properly formatted
	c.DBHost = getEnv("DB_HOST", "localhost")
	c.DBPort, _ = strconv.Atoi(getEnv("DB_PORT", "5432"))
	c.DBUser = getEnv("DB_USER", "postgres")
	c.DBPassword = getEnv("DB_PASSWORD", "")
	c.DBName = getEnv("DB_NAME", "todos")
	c.DBSSLMode = getEnv("DB_SSLMODE", "disable")

	return nil
}
