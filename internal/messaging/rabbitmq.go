package messaging

import (
	"context"
	"fmt"
	"strings"
	"subsnotifpro-go/internal/logger"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	conn      *amqp.Connection
	ch        *amqp.Channel
	connMutex sync.Mutex
	closed    chan *amqp.Error
	retrying  bool
)

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
		// 🛑 Stop reconnecting if shutting down
		if ctx.Err() != nil {
			logger.Log.Warn("⚠️ Shutdown in progress, stopping RabbitMQ reconnection")
			return fmt.Errorf("RabbitMQ shutdown in progress")
		}

		conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err == nil {
			break
		}

		logger.Log.Warnf("⚠️ RabbitMQ reconnect attempt %d failed: %v", i, err)
		time.Sleep(time.Second * 2)
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

	logger.Log.Info("✅ Connected to RabbitMQ")
	closed = make(chan *amqp.Error, 1)
	ch.NotifyClose(closed)

	go func() {
		select {
		case err := <-closed:
			if err != nil {
				logger.Log.Warn("⚠️ RabbitMQ connection lost! Reconnecting...")
				connectRabbitMQ(ctx)
			}
		case <-ctx.Done():
			logger.Log.Warn("🚦 Shutdown detected, stopping reconnection")
		}
	}()

	return nil
}

// GetChannel ensures the RabbitMQ channel is open and reconnects if broken
func GetChannel(ctx context.Context) (*amqp.Channel, error) {
	connMutex.Lock()

	// ✅ If connection is nil, release lock before reconnecting
	if conn == nil || ch == nil {
		logger.Log.Warn("⚠️ RabbitMQ connection lost! Reconnecting...")
		connMutex.Unlock() // 🚀 Release lock before calling connectRabbitMQ()
		if err := connectRabbitMQ(ctx); err != nil {
			return nil, err
		}
		connMutex.Lock() // 🔄 Re-acquire lock after reconnecting
	}

	// ✅ Check if the channel is closed using `NotifyClose`
	notifyChan := make(chan *amqp.Error, 1)
	ch.NotifyClose(notifyChan)

	select {
	case err := <-notifyChan:
		logger.Log.Warn("⚠️ RabbitMQ channel closed. Reconnecting...", err)
		connMutex.Unlock() // 🚀 Release lock before reconnecting
		if err := connectRabbitMQ(ctx); err != nil {
			return nil, err
		}
		connMutex.Lock() // 🔄 Re-acquire lock after reconnecting
	default:
		// ✅ Channel is still active
	}

	defer connMutex.Unlock()
	return ch, nil
}

// CloseRabbitMQ ensures all connections and channels are closed properly
func CloseRabbitMQ() {
	connMutex.Lock()
	defer connMutex.Unlock()

	if ch != nil {
		logger.Log.Warn("🚦 Closing RabbitMQ channel...")
		err := ch.Close()
		if err != nil && !strings.Contains(err.Error(), "channel/connection is not open") {
			logger.Log.Warn("⚠️ Error closing RabbitMQ channel:", err)
		}
		ch = nil
	}

	if conn != nil {
		logger.Log.Warn("🚦 Closing RabbitMQ connection...")
		err := conn.Close()
		if err != nil && !strings.Contains(err.Error(), "channel/connection is not open") {
			logger.Log.Warn("⚠️ Error closing RabbitMQ connection:", err)
		}
		conn = nil
	}

	// ✅ Prevent blocked channel on shutdown
	select {
	case <-closed:
	default:
		close(closed)
	}

	logger.Log.Warn("✅ RabbitMQ connection closed")
}
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
				logger.Log.Warn("⚠️ RabbitMQ connection lost! Reconnecting...")

				go func() {
					if err := connectRabbitMQ(ctx); err != nil {
						logger.Log.Warn("❌ Failed to reconnect to RabbitMQ:", err)
					}
				}()
			}
		}
	}
}
