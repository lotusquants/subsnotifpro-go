package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type RabbitMQConfig struct {
	// Connection
	Username       string
	Password       string
	Host           string
	Port           string
	VHost          string
	ConnectionName string

	// Retry Policy
	MaxRetries int
	RetryDelay time.Duration

	// Queues
	RTDN struct {
		Exchange      string
		Queue         string
		RoutingKey    string
		DLX           string
		DLQ           string
		DLQRoutingKey string
	}
	AppStore struct {
		Exchange      string
		Queue         string
		RoutingKey    string
		DLX           string
		DLQ           string
		DLQRoutingKey string
	}
	UnifiedSubs struct {
		Exchange      string
		Queue         string
		RoutingKey    string
		DLX           string
		DLQ           string
		DLQRoutingKey string
	}

	// Worker Configuration
	WorkerCount int

	// DLQ Monitoring
	DLQSizeLowThreshold    int
	DLQSizeMediumThreshold int
	DLQCheckIntervalLow    time.Duration
	DLQCheckIntervalMedium time.Duration
	DLQCheckIntervalHigh   time.Duration
}

// Config stores all environment configurations
type Config struct {
	ServerPort string
	ServerMode string

	// Database Configuration
	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// RabbitMQ Configuration
	RabbitMQ RabbitMQConfig
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

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerMode: getEnv("SERVER_MODE", "release"),

		// Database Configuration
		DBType:     getEnv("DB_TYPE", "postgres"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// RabbitMQ Configuration
	// RabbitMQ Configuration
	cfg.RabbitMQ.Username = getEnv("RABBITMQ_USERNAME", "guest")
	cfg.RabbitMQ.Password = getEnv("RABBITMQ_PASSWORD", "guest")
	cfg.RabbitMQ.Host = getEnv("RABBITMQ_HOST", "localhost")
	cfg.RabbitMQ.Port = getEnv("RABBITMQ_PORT", "5672")
	cfg.RabbitMQ.VHost = getEnv("RABBITMQ_VHOST", "/")
	cfg.RabbitMQ.ConnectionName = getEnv("RABBITMQ_CONNECTION_NAME", "subsnotifpro-go")

	// Retry Policy
	cfg.RabbitMQ.MaxRetries = getEnvAsInt("RABBITMQ_MAX_RETRIES", 3)
	cfg.RabbitMQ.RetryDelay = getEnvAsDuration("RABBITMQ_RETRY_DELAY", 5*time.Second)

	// RTDN Queue
	cfg.RabbitMQ.RTDN.Exchange = getEnv("RABBITMQ_RTDN_EXCHANGE", "rtdn_exchange")
	cfg.RabbitMQ.RTDN.Queue = getEnv("RABBITMQ_RTDN_QUEUE", "rtdn_events")
	cfg.RabbitMQ.RTDN.RoutingKey = getEnv("RABBITMQ_RTDN_ROUTING_KEY", "rtdn.event")
	cfg.RabbitMQ.RTDN.DLX = getEnv("RABBITMQ_RTDN_DLX", "rtdn_dead_letter_exchange")
	cfg.RabbitMQ.RTDN.DLQ = getEnv("RABBITMQ_RTDN_DLQ", "rtdn_dlq")
	cfg.RabbitMQ.RTDN.DLQRoutingKey = getEnv("RABBITMQ_RTDN_DLQ_ROUTING_KEY", "rtdn.dlq.event")

	// App Store Queue
	cfg.RabbitMQ.AppStore.Exchange = getEnv("RABBITMQ_APPSTORE_EXCHANGE", "appstore_exchange")
	cfg.RabbitMQ.AppStore.Queue = getEnv("RABBITMQ_APPSTORE_QUEUE", "appstore_events")
	cfg.RabbitMQ.AppStore.RoutingKey = getEnv("RABBITMQ_APPSTORE_ROUTING_KEY", "appstore.event")
	cfg.RabbitMQ.AppStore.DLX = getEnv("RABBITMQ_APPSTORE_DLX", "appstore_dead_letter_exchange")
	cfg.RabbitMQ.AppStore.DLQ = getEnv("RABBITMQ_APPSTORE_DLQ", "appstore_dlq")
	cfg.RabbitMQ.AppStore.DLQRoutingKey = getEnv("RABBITMQ_APPSTORE_DLQ_ROUTING_KEY", "appstore.dlq.event")

	// Unified Subs Queue
	cfg.RabbitMQ.UnifiedSubs.Exchange = getEnv("RABBITMQ_UNIFIED_SUBS_EXCHANGE", "unified_subscription_exchange")
	cfg.RabbitMQ.UnifiedSubs.Queue = getEnv("RABBITMQ_UNIFIED_SUBS_QUEUE", "unified_subscription_events")
	cfg.RabbitMQ.UnifiedSubs.RoutingKey = getEnv("RABBITMQ_UNIFIED_SUBS_ROUTING_KEY", "subscription.update")
	cfg.RabbitMQ.UnifiedSubs.DLX = getEnv("RABBITMQ_UNIFIED_SUBS_DLX", "unified_subscription_dlx")
	cfg.RabbitMQ.UnifiedSubs.DLQ = getEnv("RABBITMQ_UNIFIED_SUBS_DLQ", "unified_subscription_dlq")
	cfg.RabbitMQ.UnifiedSubs.DLQRoutingKey = getEnv("RABBITMQ_UNIFIED_SUBS_DLQ_ROUTING_KEY", "subscription.dlq.update")

	// Worker Configuration
	cfg.RabbitMQ.WorkerCount = getEnvAsInt("RABBITMQ_WORKER_COUNT", 3)

	// DLQ Monitoring
	cfg.RabbitMQ.DLQSizeLowThreshold = getEnvAsInt("RABBITMQ_DLQ_SIZE_LOW_THRESHOLD", 10)
	cfg.RabbitMQ.DLQSizeMediumThreshold = getEnvAsInt("RABBITMQ_DLQ_SIZE_MEDIUM_THRESHOLD", 50)
	cfg.RabbitMQ.DLQCheckIntervalLow = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_LOW", 5*time.Minute)
	cfg.RabbitMQ.DLQCheckIntervalMedium = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_MEDIUM", 1*time.Minute)
	cfg.RabbitMQ.DLQCheckIntervalHigh = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_HIGH", 30*time.Second)

	return cfg
}

// getEnv fetches the environment variable or returns a default value if missing
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// getEnvAsInt fetches the environment variable as int or returns a default value if missing
func getEnvAsInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}

// getEnvAsDuration fetches the environment variable as time.Duration or returns a default value if missing
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return result
}
