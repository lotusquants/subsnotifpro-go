// queue/rabbitmq.go
package queue

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/pkg/logger"

	"github.com/streadway/amqp"
)

type RabbitMQManager struct {
	url         string
	conn        *amqp.Connection
	channel     *amqp.Channel
	mutex       sync.RWMutex
	notifyClose chan *amqp.Error
	ctx         context.Context
	cancel      context.CancelFunc
	config      config.RabbitMQConfig
}

// Exchange Types
const (
	DirectExchange  = "direct"
	TopicExchange   = "topic"
	FanoutExchange  = "fanout"
	HeadersExchange = "headers"
)

// NewRabbitMQManager creates a new RabbitMQ manager instance
func NewRabbitMQManager(ctx context.Context, cfg config.RabbitMQConfig) *RabbitMQManager {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	ctx, cancel := context.WithCancel(ctx)

	manager := &RabbitMQManager{
		url:    connString,
		ctx:    ctx,
		cancel: cancel,
		config: cfg,
	}

	go manager.startConnectionMonitor()
	return manager
}

func (rm *RabbitMQManager) GetChannel() (*amqp.Channel, error) {
	rm.mutex.RLock()
	channel := rm.channel
	rm.mutex.RUnlock()

	if channel != nil && !isChannelClosed(channel) {
		return channel, nil
	}

	return rm.reconnect()
}

func (rm *RabbitMQManager) reconnect() (*amqp.Channel, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 1. Context check
	if err := rm.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// 2. Cleanup old connection (with brief timeout)
	go func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()

		if rm.channel != nil {
			select {
			case <-cleanupCtx.Done():
				return
			default:
				_ = rm.channel.Close()
			}
		}
		if rm.conn != nil {
			select {
			case <-cleanupCtx.Done():
				return
			default:
				_ = rm.conn.Close()
			}
		}
	}()

	// 3. Establish new connection with proper timeout
	_, dialCancel := context.WithTimeout(rm.ctx, 10*time.Second)
	defer dialCancel()

	conn, err := amqp.DialConfig(rm.url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	})
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// 4. Create channel
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("channel creation failed: %w", err)
	}

	// 5. Configure channel quality of service
	if err := channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("QoS setup failed: %w", err)
	}

	// 6. Update instance state and setup close notification
	rm.conn = conn
	rm.channel = channel
	rm.notifyClose = make(chan *amqp.Error, 1)
	rm.channel.NotifyClose(rm.notifyClose)

	return channel, nil
}

func (rm *RabbitMQManager) startConnectionMonitor() {
	for {
		select {
		case <-rm.ctx.Done():
			rm.Close()
			return
		case err := <-rm.notifyClose:
			if err != nil {
				logger.Log.Warnf("RabbitMQ connection closed: %v", err)
				if _, reconnectErr := rm.reconnect(); reconnectErr != nil {
					logger.Log.Errorf("Failed to reconnect: %v", reconnectErr)
				}
			}
		}
	}
}

func (rm *RabbitMQManager) Close() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Early return if already closed
	if rm.conn == nil && rm.channel == nil {
		return
	}

	// Cancel the context first to stop any reconnection attempts
	if rm.cancel != nil {
		rm.cancel()
	}

	// Close resources with proper error handling and timeouts
	var wg sync.WaitGroup
	wg.Add(2)

	// Close channel
	go func() {
		defer wg.Done()
		if rm.channel != nil {
			err := rm.channel.Close()
			if err != nil {
				if err == amqp.ErrClosed {
					log.Println("ℹ️ RabbitMQ channel already closed")
				} else {
					log.Printf("⚠️ Error closing channel: %v", err)
				}
			} else {
				log.Println("✅ RabbitMQ channel closed")
			}
			rm.channel = nil
		}
	}()

	// Close connection
	go func() {
		defer wg.Done()
		if rm.conn != nil {
			err := rm.conn.Close()
			if err != nil {
				if err == amqp.ErrClosed {
					log.Println("ℹ️ RabbitMQ connection already closed")
				} else {
					log.Printf("⚠️ Error closing connection: %v", err)
				}
			} else {
				log.Println("✅ RabbitMQ connection closed")
			}
			rm.conn = nil
		}
	}()

	// Wait for shutdown with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("🚦 RabbitMQ shutdown completed successfully")
	case <-time.After(5 * time.Second):
		log.Println("⚠️ Timeout waiting for RabbitMQ shutdown")
	}
}

func (rm *RabbitMQManager) GetConnection() *amqp.Connection {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.conn
}

func isChannelClosed(ch *amqp.Channel) bool {
	if ch == nil {
		return true
	}

	select {
	case _, ok := <-ch.NotifyClose(make(chan *amqp.Error, 1)):
		return !ok
	default:
		return false
	}
}

// SanitizeRabbitMQURL returns a redacted connection string for logging purposes
func (rm *RabbitMQManager) SanitizeRabbitMQURL() string {
	if rm == nil {
		return "<rabbitmq-not-configured>"
	}

	// Construct the URL from config
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		rm.config.Username,
		rm.config.Password,
		rm.config.Host,
		rm.config.Port,
		strings.TrimPrefix(rm.config.VHost, "/"),
	)

	// Redact credentials
	parts := strings.Split(url, "@")
	if len(parts) == 2 {
		return "amqp://<redacted>@" + parts[1]
	}
	return url
}
