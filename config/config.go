package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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

// MessagingType represents the messaging backend type
type MessagingType string

const (
	MessagingTypeRabbitMQ   MessagingType = "rabbitmq"
	MessagingTypeServiceBus MessagingType = "servicebus"
)

type ServiceBusConfig struct {
	// Connection
	Namespace          string
	ConnectionString   string // For development only
	UseManagedIdentity bool   // Preferred for production

	// Retry Policy
	MaxRetries int
	RetryDelay time.Duration

	// Topics and Queues
	RTDN struct {
		TopicName        string
		SubscriptionName string
		DLQName          string
	}
	AppStore struct {
		TopicName        string
		SubscriptionName string
		DLQName          string
	}
	UnifiedSubs struct {
		TopicName        string
		SubscriptionName string
		DLQName          string
	}

	// Worker Configuration
	WorkerCount int
}

// DatabaseDeploymentMode represents different database deployment options
type DatabaseDeploymentMode string

const (
	DatabaseDeploymentModeContainer DatabaseDeploymentMode = "container"
	DatabaseDeploymentModeManaged   DatabaseDeploymentMode = "managed"
	DatabaseDeploymentModeExternal  DatabaseDeploymentMode = "external"
)

// DatabaseConfig stores database configuration
type DatabaseConfig struct {
	// Deployment Mode
	DeploymentMode DatabaseDeploymentMode

	// Database Type
	Type string // postgres, mysql, sqlite

	// Connection Details
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string

	// Azure-specific settings (for managed databases)
	Azure AzureDatabaseConfig

	// Connection Pool Settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	// Additional Settings
	QueryLogging bool
	SSLRequired  bool
}

// AzureDatabaseConfig stores Azure-specific database configuration
type AzureDatabaseConfig struct {
	// Authentication method
	UseManagedIdentity bool

	// Azure Database specific settings
	ServerName string // Azure Database server name
	ResourceID string // Full resource ID for managed identity
	TenantID   string // Azure tenant ID

	// SSL and Security
	SSLMode     string // require, verify-full, verify-ca, disable
	SSLRootCert string // Path to SSL root certificate

	// Connection settings
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// Config stores all environment configurations
// JWTConfig stores JWT authentication configuration
type JWTConfig struct {
	SecretKey       string
	TokenDuration   time.Duration
	RefreshDuration time.Duration
	Issuer          string
}

type Config struct {
	ServerPort string
	ServerMode string

	// Database Configuration
	Database DatabaseConfig

	// Messaging Configuration
	MessagingType MessagingType
	RabbitMQ      RabbitMQConfig
	ServiceBus    ServiceBusConfig

	// JWT Configuration
	JWT JWTConfig
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
		Database: loadDatabaseConfig(),

		// Messaging Type Selection
		MessagingType: MessagingType(getEnv("MESSAGING_TYPE", "rabbitmq")),

		// JWT Configuration
		JWT: loadJWTConfig(),
	}

	// Load configurations based on messaging type
	cfg.loadMessagingConfig()

	return cfg
}

// loadMessagingConfig loads the appropriate messaging configuration based on MessagingType
func (c *Config) loadMessagingConfig() {
	switch c.MessagingType {
	case MessagingTypeRabbitMQ:
		c.loadRabbitMQConfig()
	case MessagingTypeServiceBus:
		c.loadServiceBusConfig()
	default:
		log.Printf("Unknown messaging type: %s, defaulting to RabbitMQ", c.MessagingType)
		c.MessagingType = MessagingTypeRabbitMQ
		c.loadRabbitMQConfig()
	}
}

// loadRabbitMQConfig loads RabbitMQ configuration
func (c *Config) loadRabbitMQConfig() {
	c.RabbitMQ.Username = getEnv("RABBITMQ_USERNAME", "guest")
	c.RabbitMQ.Password = getEnv("RABBITMQ_PASSWORD", "guest")
	c.RabbitMQ.Host = getEnv("RABBITMQ_HOST", "localhost")
	c.RabbitMQ.Port = getEnv("RABBITMQ_PORT", "5672")
	c.RabbitMQ.VHost = getEnv("RABBITMQ_VHOST", "/")
	c.RabbitMQ.ConnectionName = getEnv("RABBITMQ_CONNECTION_NAME", "subsnotifpro-go")

	// Retry Policy
	c.RabbitMQ.MaxRetries = getEnvAsInt("RABBITMQ_MAX_RETRIES", 3)
	c.RabbitMQ.RetryDelay = getEnvAsDuration("RABBITMQ_RETRY_DELAY", 5*time.Second)

	// RTDN Queue
	c.RabbitMQ.RTDN.Exchange = getEnv("RABBITMQ_RTDN_EXCHANGE", "rtdn_exchange")
	c.RabbitMQ.RTDN.Queue = getEnv("RABBITMQ_RTDN_QUEUE", "rtdn_events")
	c.RabbitMQ.RTDN.RoutingKey = getEnv("RABBITMQ_RTDN_ROUTING_KEY", "rtdn.event")
	c.RabbitMQ.RTDN.DLX = getEnv("RABBITMQ_RTDN_DLX", "rtdn_dead_letter_exchange")
	c.RabbitMQ.RTDN.DLQ = getEnv("RABBITMQ_RTDN_DLQ", "rtdn_dlq")
	c.RabbitMQ.RTDN.DLQRoutingKey = getEnv("RABBITMQ_RTDN_DLQ_ROUTING_KEY", "rtdn.dlq.event")

	// App Store Queue
	c.RabbitMQ.AppStore.Exchange = getEnv("RABBITMQ_APPSTORE_EXCHANGE", "appstore_exchange")
	c.RabbitMQ.AppStore.Queue = getEnv("RABBITMQ_APPSTORE_QUEUE", "appstore_events")
	c.RabbitMQ.AppStore.RoutingKey = getEnv("RABBITMQ_APPSTORE_ROUTING_KEY", "appstore.event")
	c.RabbitMQ.AppStore.DLX = getEnv("RABBITMQ_APPSTORE_DLX", "appstore_dead_letter_exchange")
	c.RabbitMQ.AppStore.DLQ = getEnv("RABBITMQ_APPSTORE_DLQ", "appstore_dlq")
	c.RabbitMQ.AppStore.DLQRoutingKey = getEnv("RABBITMQ_APPSTORE_DLQ_ROUTING_KEY", "appstore.dlq.event")

	// Unified Subs Queue
	c.RabbitMQ.UnifiedSubs.Exchange = getEnv("RABBITMQ_UNIFIED_SUBS_EXCHANGE", "unified_subscription_exchange")
	c.RabbitMQ.UnifiedSubs.Queue = getEnv("RABBITMQ_UNIFIED_SUBS_QUEUE", "unified_subscription_events")
	c.RabbitMQ.UnifiedSubs.RoutingKey = getEnv("RABBITMQ_UNIFIED_SUBS_ROUTING_KEY", "subscription.update")
	c.RabbitMQ.UnifiedSubs.DLX = getEnv("RABBITMQ_UNIFIED_SUBS_DLX", "unified_subscription_dlx")
	c.RabbitMQ.UnifiedSubs.DLQ = getEnv("RABBITMQ_UNIFIED_SUBS_DLQ", "unified_subscription_dlq")
	c.RabbitMQ.UnifiedSubs.DLQRoutingKey = getEnv("RABBITMQ_UNIFIED_SUBS_DLQ_ROUTING_KEY", "subscription.dlq.update")

	// Worker Configuration
	c.RabbitMQ.WorkerCount = getEnvAsInt("RABBITMQ_WORKER_COUNT", 3)

	// DLQ Monitoring
	c.RabbitMQ.DLQSizeLowThreshold = getEnvAsInt("RABBITMQ_DLQ_SIZE_LOW_THRESHOLD", 10)
	c.RabbitMQ.DLQSizeMediumThreshold = getEnvAsInt("RABBITMQ_DLQ_SIZE_MEDIUM_THRESHOLD", 50)
	c.RabbitMQ.DLQCheckIntervalLow = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_LOW", 5*time.Minute)
	c.RabbitMQ.DLQCheckIntervalMedium = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_MEDIUM", 1*time.Minute)
	c.RabbitMQ.DLQCheckIntervalHigh = getEnvAsDuration("RABBITMQ_DLQ_CHECK_INTERVAL_HIGH", 30*time.Second)
}

// loadServiceBusConfig loads Azure Service Bus configuration
func (c *Config) loadServiceBusConfig() {
	c.ServiceBus.Namespace = getEnv("SERVICEBUS_NAMESPACE", "")
	c.ServiceBus.ConnectionString = getEnv("SERVICEBUS_CONNECTION_STRING", "")
	c.ServiceBus.UseManagedIdentity = getEnvAsBool("SERVICEBUS_USE_MANAGED_IDENTITY", true)

	// Retry Policy
	c.ServiceBus.MaxRetries = getEnvAsInt("SERVICEBUS_MAX_RETRIES", 3)
	c.ServiceBus.RetryDelay = getEnvAsDuration("SERVICEBUS_RETRY_DELAY", 5*time.Second)

	// RTDN Topic
	c.ServiceBus.RTDN.TopicName = getEnv("SERVICEBUS_RTDN_TOPIC", "rtdn-events")
	c.ServiceBus.RTDN.SubscriptionName = getEnv("SERVICEBUS_RTDN_SUBSCRIPTION", "rtdn-subscription")
	c.ServiceBus.RTDN.DLQName = getEnv("SERVICEBUS_RTDN_DLQ", "rtdn-dlq")

	// App Store Topic
	c.ServiceBus.AppStore.TopicName = getEnv("SERVICEBUS_APPSTORE_TOPIC", "appstore-events")
	c.ServiceBus.AppStore.SubscriptionName = getEnv("SERVICEBUS_APPSTORE_SUBSCRIPTION", "appstore-subscription")
	c.ServiceBus.AppStore.DLQName = getEnv("SERVICEBUS_APPSTORE_DLQ", "appstore-dlq")

	// Unified Subs Topic
	c.ServiceBus.UnifiedSubs.TopicName = getEnv("SERVICEBUS_UNIFIED_SUBS_TOPIC", "unified-subscription-events")
	c.ServiceBus.UnifiedSubs.SubscriptionName = getEnv("SERVICEBUS_UNIFIED_SUBS_SUBSCRIPTION", "unified-subscription")
	c.ServiceBus.UnifiedSubs.DLQName = getEnv("SERVICEBUS_UNIFIED_SUBS_DLQ", "unified-subscription-dlq")

	// Worker Configuration
	c.ServiceBus.WorkerCount = getEnvAsInt("SERVICEBUS_WORKER_COUNT", 3)
}

// loadDatabaseConfig loads database configuration based on deployment mode
func loadDatabaseConfig() DatabaseConfig {
	deploymentMode := DatabaseDeploymentMode(getEnv("DB_DEPLOYMENT_MODE", "container"))

	// Parse connection pool settings
	maxOpenConns := 25
	if val := getEnv("DB_MAX_OPEN_CONNS", ""); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxOpenConns = parsed
		}
	}

	maxIdleConns := 5
	if val := getEnv("DB_MAX_IDLE_CONNS", ""); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxIdleConns = parsed
		}
	}

	connMaxLifetime := 15 * time.Minute
	if val := getEnv("DB_CONN_MAX_LIFETIME", ""); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			connMaxLifetime = parsed
		}
	}

	config := DatabaseConfig{
		DeploymentMode:  deploymentMode,
		Type:            getEnv("DB_TYPE", "postgres"),
		Host:            getEnv("DB_HOST", getDefaultHost(deploymentMode)),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		Name:            getEnv("DB_NAME", getDefaultDBName(deploymentMode)),
		SSLMode:         getEnv("DB_SSLMODE", getDefaultSSLMode(deploymentMode)),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		QueryLogging:    getEnv("DB_QUERY_LOGGING", "false") == "true",
		SSLRequired:     getEnv("DB_SSL_REQUIRED", "false") == "true",
	}

	// Load Azure-specific configuration if using managed deployment
	if deploymentMode == DatabaseDeploymentModeManaged {
		config.Azure = AzureDatabaseConfig{
			UseManagedIdentity: getEnv("DB_AZURE_USE_MANAGED_IDENTITY", "false") == "true",
			ServerName:         getEnv("DB_AZURE_SERVER_NAME", ""),
			ResourceID:         getEnv("DB_AZURE_RESOURCE_ID", ""),
			TenantID:           getEnv("DB_AZURE_TENANT_ID", ""),
			SSLMode:            getEnv("DB_AZURE_SSL_MODE", "require"),
			SSLRootCert:        getEnv("DB_AZURE_SSL_ROOT_CERT", ""),
		}

		// Parse Azure-specific timeouts
		if val := getEnv("DB_AZURE_CONNECT_TIMEOUT", ""); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil {
				config.Azure.ConnectTimeout = parsed
			}
		}

		if val := getEnv("DB_AZURE_READ_TIMEOUT", ""); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil {
				config.Azure.ReadTimeout = parsed
			}
		}

		if val := getEnv("DB_AZURE_WRITE_TIMEOUT", ""); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil {
				config.Azure.WriteTimeout = parsed
			}
		}
	}

	return config
}

// getDefaultHost returns the default host based on deployment mode
func getDefaultHost(mode DatabaseDeploymentMode) string {
	switch mode {
	case DatabaseDeploymentModeContainer:
		return "localhost"
	case DatabaseDeploymentModeManaged:
		return "your-server.postgres.database.azure.com"
	case DatabaseDeploymentModeExternal:
		return "localhost"
	default:
		return "localhost"
	}
}

// getDefaultDBName returns the default database name based on deployment mode
func getDefaultDBName(mode DatabaseDeploymentMode) string {
	switch mode {
	case DatabaseDeploymentModeContainer:
		return "subsnotifpro_db"
	case DatabaseDeploymentModeManaged:
		return "subsnotifpro"
	case DatabaseDeploymentModeExternal:
		return "postgres"
	default:
		return "postgres"
	}
}

// getDefaultSSLMode returns the default SSL mode based on deployment mode
func getDefaultSSLMode(mode DatabaseDeploymentMode) string {
	switch mode {
	case DatabaseDeploymentModeContainer:
		return "disable"
	case DatabaseDeploymentModeManaged:
		return "require"
	case DatabaseDeploymentModeExternal:
		return "disable"
	default:
		return "disable"
	}
}

// getEnvAsBool retrieves an environment variable as a boolean value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
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

// loadJWTConfig loads JWT configuration from environment variables
func loadJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:       getEnv("JWT_SECRET_KEY", "your-secret-key-change-this-in-production"),
		TokenDuration:   getEnvAsDuration("JWT_TOKEN_DURATION", 24*time.Hour),
		RefreshDuration: getEnvAsDuration("JWT_REFRESH_DURATION", 7*24*time.Hour),
		Issuer:          getEnv("JWT_ISSUER", "subsnotifpro-go"),
	}
}

// getEnvAsDuration fetches the environment variable as duration or returns a default value if missing
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
