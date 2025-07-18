package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConnectivityTest tests database connectivity
type DatabaseConnectivityTest struct {
	config *DatabaseConfig
}

// DatabaseConfig holds database configuration for testing
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewDatabaseConnectivityTest creates a new database connectivity test
func NewDatabaseConnectivityTest() *DatabaseConnectivityTest {
	return &DatabaseConnectivityTest{
		config: &DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			Database: getEnvOrDefault("DB_NAME", "subsnotifpro_test"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
	}
}

// TestConnection tests database connection
func (d *DatabaseConnectivityTest) TestConnection(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		d.config.Host, d.config.User, d.config.Password, d.config.Database, d.config.Port, d.config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	t.Log("Database connection successful")
}

// TestQuery tests database query execution
func (d *DatabaseConnectivityTest) TestQuery(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		d.config.Host, d.config.User, d.config.Password, d.config.Database, d.config.Port, d.config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Test simple query
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute test query: %v", err)
	}

	if result != 1 {
		t.Fatalf("Expected result 1, got %d", result)
	}

	t.Log("Database query test successful")
}

// TestConnectionPool tests database connection pooling
func (d *DatabaseConnectivityTest) TestConnectionPool(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		d.config.Host, d.config.User, d.config.Password, d.config.Database, d.config.Port, d.config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Configure connection pool
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test multiple concurrent connections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		go func(id int) {
			var result int
			err := db.WithContext(ctx).Raw("SELECT $1", id).Scan(&result).Error
			if err != nil {
				t.Errorf("Concurrent query %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait a bit for concurrent operations
	time.Sleep(2 * time.Second)

	// Check pool stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		t.Error("No open connections in pool")
	}

	t.Logf("Connection pool test successful. Open connections: %d", stats.OpenConnections)
}

// RabbitMQConnectivityTest tests RabbitMQ connectivity
type RabbitMQConnectivityTest struct {
	config *RabbitMQConfig
}

// RabbitMQConfig holds RabbitMQ configuration for testing
type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Vhost    string
}

// NewRabbitMQConnectivityTest creates a new RabbitMQ connectivity test
func NewRabbitMQConnectivityTest() *RabbitMQConnectivityTest {
	return &RabbitMQConnectivityTest{
		config: &RabbitMQConfig{
			Host:     getEnvOrDefault("RABBITMQ_HOST", "localhost"),
			Port:     getEnvOrDefault("RABBITMQ_PORT", "5672"),
			Username: getEnvOrDefault("RABBITMQ_USERNAME", "guest"),
			Password: getEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
			Vhost:    getEnvOrDefault("RABBITMQ_VHOST", "/"),
		},
	}
}

// TestConnection tests RabbitMQ connection
func (r *RabbitMQConnectivityTest) TestConnection(t *testing.T) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.config.Username, r.config.Password, r.config.Host, r.config.Port, r.config.Vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	if conn.IsClosed() {
		t.Fatal("RabbitMQ connection is closed")
	}

	t.Log("RabbitMQ connection successful")
}

// TestChannel tests RabbitMQ channel operations
func (r *RabbitMQConnectivityTest) TestChannel(t *testing.T) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.config.Username, r.config.Password, r.config.Host, r.config.Port, r.config.Vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Test queue declaration
	queueName := "test_connectivity_queue"
	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	// Test message publishing
	err = ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("connectivity test message"),
		})
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	t.Log("RabbitMQ channel operations successful")
}

// TestConsumer tests RabbitMQ consumer setup
func (r *RabbitMQConnectivityTest) TestConsumer(t *testing.T) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		r.config.Username, r.config.Password, r.config.Host, r.config.Port, r.config.Vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Test queue declaration
	queueName := "test_consumer_queue"
	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	// Test consumer setup
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		t.Fatalf("Failed to register consumer: %v", err)
	}

	// Test message publishing and consumption
	go func() {
		err = ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("test consumer message"),
			})
		if err != nil {
			t.Errorf("Failed to publish test message: %v", err)
		}
	}()

	// Wait for message
	select {
	case msg := <-msgs:
		if string(msg.Body) != "test consumer message" {
			t.Errorf("Expected 'test consumer message', got '%s'", string(msg.Body))
		}
		t.Log("RabbitMQ consumer test successful")
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

// ServiceBusConnectivityTest tests Azure Service Bus connectivity
type ServiceBusConnectivityTest struct {
	connectionString string
}

// NewServiceBusConnectivityTest creates a new Service Bus connectivity test
func NewServiceBusConnectivityTest() *ServiceBusConnectivityTest {
	return &ServiceBusConnectivityTest{
		connectionString: getEnvOrDefault("SERVICEBUS_CONNECTION_STRING", ""),
	}
}

// TestConnection tests Service Bus connection
func (s *ServiceBusConnectivityTest) TestConnection(t *testing.T) {
	if s.connectionString == "" {
		t.Skip("SERVICEBUS_CONNECTION_STRING not set, skipping Service Bus tests")
	}

	client, err := azservicebus.NewClientFromConnectionString(s.connectionString, nil)
	if err != nil {
		t.Fatalf("Failed to create Service Bus client: %v", err)
	}
	defer client.Close(context.Background())

	t.Log("Service Bus connection successful")
}

// TestSender tests Service Bus sender operations
func (s *ServiceBusConnectivityTest) TestSender(t *testing.T) {
	if s.connectionString == "" {
		t.Skip("SERVICEBUS_CONNECTION_STRING not set, skipping Service Bus tests")
	}

	client, err := azservicebus.NewClientFromConnectionString(s.connectionString, nil)
	if err != nil {
		t.Fatalf("Failed to create Service Bus client: %v", err)
	}
	defer client.Close(context.Background())

	queueName := "test-connectivity-queue"
	sender, err := client.NewSender(queueName, nil)
	if err != nil {
		t.Fatalf("Failed to create sender: %v", err)
	}
	defer sender.Close(context.Background())

	// Test message sending
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := &azservicebus.Message{
		Body: []byte("connectivity test message"),
	}

	err = sender.SendMessage(ctx, message, nil)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	t.Log("Service Bus sender test successful")
}

// TestReceiver tests Service Bus receiver operations
func (s *ServiceBusConnectivityTest) TestReceiver(t *testing.T) {
	if s.connectionString == "" {
		t.Skip("SERVICEBUS_CONNECTION_STRING not set, skipping Service Bus tests")
	}

	client, err := azservicebus.NewClientFromConnectionString(s.connectionString, nil)
	if err != nil {
		t.Fatalf("Failed to create Service Bus client: %v", err)
	}
	defer client.Close(context.Background())

	queueName := "test-connectivity-queue"
	
	// Create sender to send test message
	sender, err := client.NewSender(queueName, nil)
	if err != nil {
		t.Fatalf("Failed to create sender: %v", err)
	}
	defer sender.Close(context.Background())

	// Send test message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := &azservicebus.Message{
		Body: []byte("test receiver message"),
	}

	err = sender.SendMessage(ctx, message, nil)
	if err != nil {
		t.Fatalf("Failed to send test message: %v", err)
	}

	// Create receiver
	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		t.Fatalf("Failed to create receiver: %v", err)
	}
	defer receiver.Close(context.Background())

	// Receive message
	messages, err := receiver.ReceiveMessages(ctx, 1, nil)
	if err != nil {
		t.Fatalf("Failed to receive messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if string(messages[0].Body) != "test receiver message" {
		t.Errorf("Expected 'test receiver message', got '%s'", string(messages[0].Body))
	}

	// Complete the message
	err = receiver.CompleteMessage(ctx, messages[0], nil)
	if err != nil {
		t.Errorf("Failed to complete message: %v", err)
	}

	t.Log("Service Bus receiver test successful")
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunAllConnectivityTests runs all connectivity tests
func RunAllConnectivityTests(t *testing.T) {
	t.Run("Database", func(t *testing.T) {
		dbTest := NewDatabaseConnectivityTest()
		t.Run("Connection", dbTest.TestConnection)
		t.Run("Query", dbTest.TestQuery)
		t.Run("ConnectionPool", dbTest.TestConnectionPool)
	})

	t.Run("RabbitMQ", func(t *testing.T) {
		if os.Getenv("MESSAGING_TYPE") != "rabbitmq" {
			t.Skip("MESSAGING_TYPE is not rabbitmq, skipping RabbitMQ tests")
		}
		
		rbTest := NewRabbitMQConnectivityTest()
		t.Run("Connection", rbTest.TestConnection)
		t.Run("Channel", rbTest.TestChannel)
		t.Run("Consumer", rbTest.TestConsumer)
	})

	t.Run("ServiceBus", func(t *testing.T) {
		if os.Getenv("MESSAGING_TYPE") != "servicebus" {
			t.Skip("MESSAGING_TYPE is not servicebus, skipping Service Bus tests")
		}
		
		sbTest := NewServiceBusConnectivityTest()
		t.Run("Connection", sbTest.TestConnection)
		t.Run("Sender", sbTest.TestSender)
		t.Run("Receiver", sbTest.TestReceiver)
	})
}

// TestMain runs connectivity tests
func TestMain(m *testing.M) {
	log.Println("Running connectivity tests...")
	
	// Set up test environment
	os.Setenv("DB_NAME", "subsnotifpro_test")
	
	// Run tests
	code := m.Run()
	
	// Clean up if needed
	
	os.Exit(code)
}
