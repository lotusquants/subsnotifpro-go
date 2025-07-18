package messaging

import (
	"context"
	"fmt"
	"subsnotifpro-go/config"
	"time"

	"github.com/streadway/amqp"
)

// MessagingFactory creates publishers and consumers based on configuration
type MessagingFactory struct {
	config *config.Config
}

// NewMessagingFactory creates a new messaging factory
func NewMessagingFactory(cfg *config.Config) *MessagingFactory {
	return &MessagingFactory{config: cfg}
}

// CreatePublisher creates a publisher based on the configured messaging type
func (f *MessagingFactory) CreatePublisher(opts interface{}) (MessagePublisher, error) {
	switch f.config.MessagingType {
	case config.MessagingTypeRabbitMQ:
		// Cast options to RabbitMQ publisher options
		rabbitOpts, ok := opts.(PublisherOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for RabbitMQ publisher")
		}
		return NewRabbitMQPublisher(rabbitOpts), nil

	case config.MessagingTypeServiceBus:
		// Cast options to Service Bus publisher options
		serviceBusOpts, ok := opts.(ServiceBusPublisherOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for Service Bus publisher")
		}
		return NewServiceBusPublisher(serviceBusOpts)

	default:
		return nil, fmt.Errorf("unsupported messaging type: %s", f.config.MessagingType)
	}
}

// CreateConsumer creates a consumer based on the configured messaging type
func (f *MessagingFactory) CreateConsumer(opts interface{}) (MessageConsumer, error) {
	switch f.config.MessagingType {
	case config.MessagingTypeRabbitMQ:
		// Cast options to RabbitMQ consumer options
		rabbitOpts, ok := opts.(RabbitMQConsumerOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for RabbitMQ consumer")
		}
		return NewRabbitMQConsumer(rabbitOpts), nil

	case config.MessagingTypeServiceBus:
		// Cast options to Service Bus consumer options
		serviceBusOpts, ok := opts.(ServiceBusConsumerOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for Service Bus consumer")
		}
		return NewServiceBusConsumer(serviceBusOpts)

	default:
		return nil, fmt.Errorf("unsupported messaging type: %s", f.config.MessagingType)
	}
}

// MessageConsumer interface for both RabbitMQ and Service Bus consumers
type MessageConsumer interface {
	Start(ctx context.Context)
	Close() error
}

// RabbitMQConsumerOptions wraps the existing Consumer struct
type RabbitMQConsumerOptions struct {
	Channel      *amqp.Channel
	ExchangeName string
	RoutingKey   string
	QueueName    string
	DLQName      string
	MaxRetries   int
	WorkerCount  int
	Handler      func(ctx context.Context, payload []byte) error
	Publisher    MessagePublisher
}

// NewRabbitMQConsumer creates a RabbitMQ consumer using the existing Consumer struct
func NewRabbitMQConsumer(opts RabbitMQConsumerOptions) MessageConsumer {
	consumer := NewConsumer(
		opts.Channel,
		opts.ExchangeName,
		opts.RoutingKey,
		opts.QueueName,
		opts.DLQName,
		opts.MaxRetries,
		opts.WorkerCount,
		opts.Handler,
		opts.Publisher,
	)
	return &RabbitMQConsumerWrapper{consumer: consumer}
}

// RabbitMQConsumerWrapper wraps the existing Consumer to implement MessageConsumer interface
type RabbitMQConsumerWrapper struct {
	consumer *Consumer
}

func (w *RabbitMQConsumerWrapper) Start(ctx context.Context) {
	w.consumer.Start(ctx)
}

func (w *RabbitMQConsumerWrapper) Close() error {
	// The existing Consumer struct doesn't have a Close method
	// So we'll just return nil for now - this could be enhanced later
	return nil
}

// GetPublisherOptions returns the appropriate publisher options based on messaging type
func (f *MessagingFactory) GetPublisherOptions(channel *amqp.Channel, topicName string) interface{} {
	switch f.config.MessagingType {
	case config.MessagingTypeRabbitMQ:
		return PublisherOptions{
			Channel: channel,
		}

	case config.MessagingTypeServiceBus:
		return ServiceBusPublisherOptions{
			Namespace:          f.config.ServiceBus.Namespace,
			ConnectionString:   f.config.ServiceBus.ConnectionString,
			UseManagedIdentity: f.config.ServiceBus.UseManagedIdentity,
			TopicName:          topicName,
		}

	default:
		return nil
	}
}

// GetConsumerOptions returns the appropriate consumer options based on messaging type
func (f *MessagingFactory) GetConsumerOptions(
	channel *amqp.Channel,
	topicName, subscriptionName string,
	handler func(ctx context.Context, payload []byte) error,
	publisher MessagePublisher,
) interface{} {
	switch f.config.MessagingType {
	case config.MessagingTypeRabbitMQ:
		// Map Service Bus topic/subscription to RabbitMQ exchange/queue
		var exchangeName, queueName string

		if topicName == f.config.ServiceBus.RTDN.TopicName {
			exchangeName = f.config.RabbitMQ.RTDN.Exchange
			queueName = f.config.RabbitMQ.RTDN.Queue
		} else if topicName == f.config.ServiceBus.AppStore.TopicName {
			exchangeName = f.config.RabbitMQ.AppStore.Exchange
			queueName = f.config.RabbitMQ.AppStore.Queue
		} else if topicName == f.config.ServiceBus.UnifiedSubs.TopicName {
			exchangeName = f.config.RabbitMQ.UnifiedSubs.Exchange
			queueName = f.config.RabbitMQ.UnifiedSubs.Queue
		}

		return RabbitMQConsumerOptions{
			Channel:      channel,
			ExchangeName: exchangeName,
			RoutingKey:   queueName, // Use queue name as routing key
			QueueName:    queueName,
			DLQName:      queueName + "_dlq",
			MaxRetries:   f.config.RabbitMQ.MaxRetries,
			WorkerCount:  f.config.RabbitMQ.WorkerCount,
			Handler:      handler,
			Publisher:    publisher,
		}

	case config.MessagingTypeServiceBus:
		return ServiceBusConsumerOptions{
			Namespace:          f.config.ServiceBus.Namespace,
			ConnectionString:   f.config.ServiceBus.ConnectionString,
			UseManagedIdentity: f.config.ServiceBus.UseManagedIdentity,
			TopicName:          topicName,
			SubscriptionName:   subscriptionName,
			MaxRetries:         f.config.ServiceBus.MaxRetries,
			WorkerCount:        f.config.ServiceBus.WorkerCount,
			Handler:            handler,
			Publisher:          publisher,
		}

	default:
		return nil
	}
}

// Convenience methods for creating publishers and consumers

// NewPublisher creates a publisher based on the configuration
func NewPublisher(cfg *config.Config) (MessagePublisher, error) {
	factory := NewMessagingFactory(cfg)

	switch cfg.MessagingType {
	case config.MessagingTypeRabbitMQ:
		// For RabbitMQ, we need to create a connection and channel
		// This is a simple approach - in production, you might want to manage connections differently
		rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
			cfg.RabbitMQ.Username,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
			cfg.RabbitMQ.VHost,
		)

		conn, err := amqp.Dial(rabbitURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
		}

		opts := PublisherOptions{
			Channel: ch,
		}

		publisher, err := factory.CreatePublisher(opts)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}

		// Wrap the publisher to handle connection cleanup
		return &PublisherWrapper{
			Publisher:  publisher,
			Channel:    ch,
			Connection: conn,
		}, nil

	case config.MessagingTypeServiceBus:
		opts := ServiceBusPublisherOptions{
			Namespace:          cfg.ServiceBus.Namespace,
			ConnectionString:   cfg.ServiceBus.ConnectionString,
			UseManagedIdentity: cfg.ServiceBus.UseManagedIdentity,
		}

		return factory.CreatePublisher(opts)

	default:
		return nil, fmt.Errorf("unsupported messaging type: %s", cfg.MessagingType)
	}
}

// NewMessageConsumer creates a consumer based on the configuration
func NewMessageConsumer(cfg *config.Config) (MessageConsumer, error) {
	// For consumers, we need specific topic/queue details
	// This would typically be called with specific parameters
	return nil, fmt.Errorf("consumer creation requires specific topic/queue details - use CreateConsumer with appropriate options")
}

// PublisherWrapper wraps a publisher to handle connection cleanup
type PublisherWrapper struct {
	Publisher  MessagePublisher
	Channel    *amqp.Channel
	Connection *amqp.Connection
}

func (pw *PublisherWrapper) PublishToQueue(ctx context.Context, queueName string, event interface{}) error {
	return pw.Publisher.PublishToQueue(ctx, queueName, event)
}

func (pw *PublisherWrapper) PublishToExchange(ctx context.Context, exchange, routingKey string, event interface{}) error {
	return pw.Publisher.PublishToExchange(ctx, exchange, routingKey, event)
}

func (pw *PublisherWrapper) PublishWithDelay(ctx context.Context, exchange, routingKey string, event interface{}, delay time.Duration) error {
	return pw.Publisher.PublishWithDelay(ctx, exchange, routingKey, event, delay)
}

func (pw *PublisherWrapper) Close() error {
	var err error
	if pw.Channel != nil {
		if chErr := pw.Channel.Close(); chErr != nil {
			err = chErr
		}
	}
	if pw.Connection != nil {
		if connErr := pw.Connection.Close(); connErr != nil {
			if err == nil {
				err = connErr
			}
		}
	}
	return err
}
