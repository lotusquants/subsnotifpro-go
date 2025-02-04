package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config stores all environment configurations
type Config struct {
	ServerPort string
	ServerMode string

	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// LoadConfig loads the environment variables from .env (if available) and system environment variables
func LoadConfig() *Config {
	// Load .env file only in local development
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: No .env file found. Using system environment variables.")
		}
	}

	config := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerMode: getEnv("SERVER_MODE", "release"),

		DBType:     getEnv("DB_TYPE", "postgres"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	return config
}

// getEnv fetches the environment variable or returns a default value if missing
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
