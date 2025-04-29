package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// InitializeRabbitMQ sets up all required RabbitMQ infrastructure
func (r *RabbitMQManager) InitializeRabbitMQ(ctx context.Context, ch *amqp.Channel) error {
	const maxRetries = 3
	var err error

	if ch == nil {
		return fmt.Errorf("cannot initialize RabbitMQ with nil channel")
	}

	log.Println("⚙️ Initializing RabbitMQ infrastructure...")

	for i := 0; i < maxRetries; i++ {
		log.Printf("🐇 Attempting RabbitMQ setup (attempt %d/%d)", i+1, maxRetries)
		err = r.setupRabbitMQInfrastructure(ch)
		if err == nil {
			log.Println("✅ RabbitMQ infrastructure fully initialized")
			return nil
		}

		log.Printf("⚠️ Setup failed: %v", err)
		if i < maxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying in %v...", backoff)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
}

func (r *RabbitMQManager) setupRabbitMQInfrastructure(ch *amqp.Channel) error {
	// Test channel functionality first
	if err := r.testChannel(ch); err != nil {
		return err
	}

	// Setup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Execute setup operations
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.performSetup(ch)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("RabbitMQ setup timed out after 15 seconds")
	}
}

func (r *RabbitMQManager) performSetup(ch *amqp.Channel) error {
	// 1. Declare Main Exchanges
	if err := r.declareExchange(ch, r.config.RTDN.Exchange, DirectExchange); err != nil {
		return err
	}
	if err := r.declareExchange(ch, r.config.AppStore.Exchange, DirectExchange); err != nil {
		return err
	}
	if err := r.declareExchange(ch, r.config.UnifiedSubs.Exchange, DirectExchange); err != nil {
		return err
	}

	// 2. Declare Dead Letter Exchanges
	if err := r.declareExchange(ch, r.config.RTDN.DLX, DirectExchange); err != nil {
		return err
	}
	if err := r.declareExchange(ch, r.config.AppStore.DLX, DirectExchange); err != nil {
		return err
	}
	if err := r.declareExchange(ch, r.config.UnifiedSubs.DLX, DirectExchange); err != nil {
		return err
	}

	// 3. Declare Dead Letter Queues
	if _, err := r.SafeQueueDeclare(r.config.RTDN.DLQ, nil); err != nil {
		return fmt.Errorf("RTDN DLQ declaration failed: %w", err)
	}
	if _, err := r.SafeQueueDeclare(r.config.AppStore.DLQ, nil); err != nil {
		return fmt.Errorf("AppStore DLQ declaration failed: %w", err)
	}
	if _, err := r.SafeQueueDeclare(r.config.UnifiedSubs.DLQ, nil); err != nil {
		return fmt.Errorf("UnifiedSubs DLQ declaration failed: %w", err)
	}

	// 4. Bind DLQs to DLXs
	if err := r.bindQueue(ch, r.config.RTDN.DLQ, r.config.RTDN.DLQRoutingKey, r.config.RTDN.DLX); err != nil {
		return err
	}
	if err := r.bindQueue(ch, r.config.AppStore.DLQ, r.config.AppStore.DLQRoutingKey, r.config.AppStore.DLX); err != nil {
		return err
	}
	if err := r.bindQueue(ch, r.config.UnifiedSubs.DLQ, r.config.UnifiedSubs.DLQRoutingKey, r.config.UnifiedSubs.DLX); err != nil {
		return err
	}

	// 5. Declare Main Queues with DLX policies
	rtdnArgs := amqp.Table{
		"x-dead-letter-exchange":    r.config.RTDN.DLX,
		"x-dead-letter-routing-key": r.config.RTDN.DLQRoutingKey,
	}
	if _, err := r.SafeQueueDeclare(r.config.RTDN.Queue, rtdnArgs); err != nil {
		return fmt.Errorf("failed to setup RTDN queue: %w", err)
	}

	appStoreArgs := amqp.Table{
		"x-dead-letter-exchange":    r.config.AppStore.DLX,
		"x-dead-letter-routing-key": r.config.AppStore.DLQRoutingKey,
	}
	if _, err := r.SafeQueueDeclare(r.config.AppStore.Queue, appStoreArgs); err != nil {
		return fmt.Errorf("failed to setup AppStore queue: %w", err)
	}

	unifiedSubsArgs := amqp.Table{
		"x-dead-letter-exchange":    r.config.UnifiedSubs.DLX,
		"x-dead-letter-routing-key": r.config.UnifiedSubs.DLQRoutingKey,
	}
	if _, err := r.SafeQueueDeclare(r.config.UnifiedSubs.Queue, unifiedSubsArgs); err != nil {
		return fmt.Errorf("failed to setup UnifiedSubs queue: %w", err)
	}

	// 6. Bind Main Queues to Exchanges
	if err := r.bindQueue(ch, r.config.RTDN.Queue, r.config.RTDN.RoutingKey, r.config.RTDN.Exchange); err != nil {
		return err
	}
	if err := r.bindQueue(ch, r.config.AppStore.Queue, r.config.AppStore.RoutingKey, r.config.AppStore.Exchange); err != nil {
		return err
	}
	if err := r.bindQueue(ch, r.config.UnifiedSubs.Queue, r.config.UnifiedSubs.RoutingKey, r.config.UnifiedSubs.Exchange); err != nil {
		return err
	}

	log.Println("✓ All queues and exchanges properly configured")
	return nil
}

func (r *RabbitMQManager) testChannel(ch *amqp.Channel) error {
	testQueue := "connection_test_queue_" + time.Now().Format("20060102150405")
	_, err := ch.QueueDeclare(
		testQueue,
		false, // non-durable
		true,  // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("basic channel test failed: %w", err)
	}

	// Clean up test queue
	_, err = ch.QueueDelete(testQueue, false, false, false)
	if err != nil {
		return fmt.Errorf("test queue cleanup failed: %w", err)
	}

	log.Println("✓ Basic channel test passed")
	return nil
}

func (r *RabbitMQManager) declareExchange(ch *amqp.Channel, name, kind string) error {
	err := ch.ExchangeDeclare(
		name,
		kind,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", name, err)
	}
	log.Printf("✓ Exchange %s (%s) declared", name, kind)
	return nil
}

func (r *RabbitMQManager) bindQueue(ch *amqp.Channel, queue, routingKey, exchange string) error {
	err := ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s with key %s: %w",
			queue, exchange, routingKey, err)
	}
	log.Printf("✓ Queue %s bound to exchange %s with key %s", queue, exchange, routingKey)
	return nil
}

func (r *RabbitMQManager) SafeQueueDeclare(name string, args amqp.Table) (amqp.Queue, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log.Printf("✓ Safe creating queue %s", name)

	for i := 0; i < 3; i++ {
		if r.conn.IsClosed() {
			newCh, err := r.conn.Channel()
			if err != nil {
				return amqp.Queue{}, fmt.Errorf("channel recovery failed: %w", err)
			}
			r.channel = newCh
		}

		q, err := r.channel.QueueDeclare(
			name,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			args,
		)
		if err == nil {
			log.Printf("✓ Queue %s created", name)
			return q, nil
		}

		if shouldRetry(err) {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return amqp.Queue{}, err
	}

	return amqp.Queue{}, fmt.Errorf("max retries exceeded for queue %s", name)
}

func shouldRetry(err error) bool {
	e, ok := err.(*amqp.Error)
	if !ok {
		return false
	}
	// Retry on these error codes
	return e.Code == 504 || e.Code == 506 || e.Code == 541
}
