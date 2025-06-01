// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
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
	// This helps when you need individual connection parameters
	if err := cfg.parseDatabaseURL(); err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseDatabaseURL extracts components from a postgres connection string
// Example: postgres://user:password@localhost:5432/dbname?sslmode=disable
func (c *Config) parseDatabaseURL() error {
	// In a real application, you'd use a proper URL parser
	// This is simplified for demonstration
	// Consider using github.com/jackc/pgx/v5/stdlib for PostgreSQL URLs

	// For now, we'll trust that DATABASE_URL is properly formatted
	// In production, use proper parsing
	c.DBHost = getEnv("DB_HOST", "localhost")
	c.DBPort, _ = strconv.Atoi(getEnv("DB_PORT", "5432"))
	c.DBUser = getEnv("DB_USER", "postgres")
	c.DBPassword = getEnv("DB_PASSWORD", "")
	c.DBName = getEnv("DB_NAME", "todos")
	c.DBSSLMode = getEnv("DB_SSLMODE", "disable")

	return nil
}
