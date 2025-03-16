package messaging

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"subsnotifpro-go/internal/logger"

	"github.com/streadway/amqp"
)

var (
	conn      *amqp.Connection
	ch        *amqp.Channel
	connMutex sync.Mutex
	closed    chan *amqp.Error
	retrying  bool
)

var (
	adminMutex  sync.Mutex
	rabbitmqURL string
)

func init() {
	rabbitmqURL = getRabbitMQURL()
	logger.Log.Infof("🚀 RabbitMQ URL: %s", sanitizeRabbitMQURL(rabbitmqURL))
}

// getRabbitMQURL fetches RabbitMQ URL from env, with fallback to default.
func getRabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost:5672/" // Fallback
}

// sanitizeRabbitMQURL removes credentials before logging URL.
func sanitizeRabbitMQURL(url string) string {
	parts := strings.Split(url, "@")
	if len(parts) == 2 {
		return "amqp://<redacted>@" + parts[1]
	}
	return url
}

func connectRabbitMQ(ctx context.Context) error {
	connMutex.Lock()
	defer connMutex.Unlock()

	if retrying {
		logger.Log.Warn("⚠️ Already reconnecting RabbitMQ, skipping duplicate request...")
		return nil
	}

	retrying = true
	defer func() { retrying = false }()

	var err error
	for i := 1; i <= 5; i++ {
		if ctx.Err() != nil {
			logger.Log.Warn("⚠️ Shutdown in progress, stopping RabbitMQ reconnection")
			return fmt.Errorf("RabbitMQ shutdown in progress")
		}

		logger.Log.Infof("🔄 Attempting RabbitMQ connection (attempt %d/5)...", i)

		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			logger.Log.Info("✅ RabbitMQ connected successfully.")
			break
		}

		logger.Log.Warnf("⚠️ RabbitMQ reconnect attempt %d failed: %v", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Log.Fatal("❌ RabbitMQ connection failed after retries. Ensure RabbitMQ is running!")
		return err
	}

	ch, err = conn.Channel()
	if err != nil {
		logger.Log.Fatal("❌ Failed to open RabbitMQ channel:", err)
		return err
	}

	logger.Log.Info("✅ RabbitMQ channel opened successfully.")

	closed = make(chan *amqp.Error, 1)
	ch.NotifyClose(closed)

	go monitorConnectionClosure(ctx)

	return nil
}

// monitorConnectionClosure listens for RabbitMQ connection closure.
func monitorConnectionClosure(ctx context.Context) {
	select {
	case err := <-closed:
		if err != nil {
			logger.Log.Warn("⚠️ RabbitMQ connection lost! Reconnecting...")
			connectRabbitMQ(ctx)
		}
	case <-ctx.Done():
		logger.Log.Warn("🚦 Shutdown detected, stopping reconnection monitoring")
	}
}

// GetChannel ensures RabbitMQ connection is healthy and reconnects if needed.
func GetChannel(ctx context.Context) (*amqp.Channel, error) {
	connMutex.Lock()

	if conn == nil || ch == nil {
		logger.Log.Warn("⚠️ RabbitMQ connection lost! Reconnecting...")
		connMutex.Unlock()
		if err := connectRabbitMQ(ctx); err != nil {
			return nil, err
		}
		connMutex.Lock()
	}

	notifyChan := make(chan *amqp.Error, 1)
	ch.NotifyClose(notifyChan)

	select {
	case err := <-notifyChan:
		logger.Log.Warnf("⚠️ RabbitMQ channel closed unexpectedly. Reconnecting...: %v", err)
		connMutex.Unlock()
		if err := connectRabbitMQ(ctx); err != nil {
			return nil, err
		}
		connMutex.Lock()
	default:
		// Channel is healthy.
	}

	defer connMutex.Unlock()
	return ch, nil
}

// CloseRabbitMQ closes RabbitMQ connection and channel safely.
func CloseRabbitMQ() {
	connMutex.Lock()
	defer connMutex.Unlock()

	if ch != nil {
		logger.Log.Warn("🚦 Closing RabbitMQ channel...")
		if err := ch.Close(); err != nil && !strings.Contains(err.Error(), "channel/connection is not open") {
			logger.Log.Warnf("⚠️ Error closing RabbitMQ channel: %v", err)
		}
		ch = nil
	}

	if conn != nil {
		logger.Log.Warn("🚦 Closing RabbitMQ connection...")
		if err := conn.Close(); err != nil && !strings.Contains(err.Error(), "channel/connection is not open") {
			logger.Log.Warnf("⚠️ Error closing RabbitMQ connection: %v", err)
		}
		conn = nil
	}

	select {
	case <-closed:
	default:
		close(closed)
	}

	logger.Log.Warn("✅ RabbitMQ connection fully closed.")
}

// MonitorRabbitMQConnection periodically checks RabbitMQ health and triggers reconnect if needed.
func MonitorRabbitMQConnection(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Warn("🚦 Stopping RabbitMQ connection monitoring...")
			return
		case <-ticker.C:
			connMutex.Lock()
			needReconnect := (conn == nil || ch == nil) && !retrying
			connMutex.Unlock()

			if needReconnect {
				logger.Log.Warn("⚠️ RabbitMQ connection lost! Triggering reconnect...")
				go func() {
					if err := connectRabbitMQ(ctx); err != nil {
						logger.Log.Warnf("❌ Failed to reconnect RabbitMQ: %v", err)
					}
				}()
			}
		}
	}
}

// GetAdminChannel provides a short-lived connection and channel for admin tasks like inspecting the DLQ.
func GetAdminChannel() (*amqp.Connection, *amqp.Channel, error) {
	adminMutex.Lock()
	defer adminMutex.Unlock()

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		logger.Log.Errorf("❌ Failed to create admin RabbitMQ connection: %v", err)
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Log.Warn("⚠️ Closed admin connection after failing to create channel.")
		logger.Log.Errorf("❌ Failed to create admin RabbitMQ channel: %v", err)
		return nil, nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	logger.Log.Debug("✅ Created short-lived RabbitMQ admin connection and channel")
	return conn, ch, nil
}
