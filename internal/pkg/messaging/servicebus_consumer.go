package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ServiceBusConsumer implements a consumer for Azure Service Bus
type ServiceBusConsumer struct {
	client           *azservicebus.Client
	receiver         *azservicebus.Receiver
	topicName        string
	subscriptionName string
	maxRetries       int
	workerCount      int
	handler          func(ctx context.Context, payload []byte) error
	publisher        MessagePublisher
}

// ServiceBusConsumerOptions contains configuration for ServiceBus consumer
type ServiceBusConsumerOptions struct {
	Namespace         string
	ConnectionString  string
	UseManagedIdentity bool
	TopicName         string
	SubscriptionName  string
	MaxRetries        int
	WorkerCount       int
	Handler           func(ctx context.Context, payload []byte) error
	Publisher         MessagePublisher
}

// NewServiceBusConsumer creates a new Azure Service Bus consumer
func NewServiceBusConsumer(opts ServiceBusConsumerOptions) (*ServiceBusConsumer, error) {
	var client *azservicebus.Client
	var err error

	if opts.UseManagedIdentity {
		// Use Managed Identity (recommended for Azure environments)
		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create managed identity credential: %w", err)
		}
		
		fullyQualifiedNamespace := fmt.Sprintf("%s.servicebus.windows.net", opts.Namespace)
		client, err = azservicebus.NewClient(fullyQualifiedNamespace, credential, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create service bus client with managed identity: %w", err)
		}
	} else {
		// Use connection string (for development)
		client, err = azservicebus.NewClientFromConnectionString(opts.ConnectionString, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create service bus client from connection string: %w", err)
		}
	}

	// Create receiver for the subscription
	receiver, err := client.NewReceiverForSubscription(opts.TopicName, opts.SubscriptionName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create receiver for topic %s, subscription %s: %w", opts.TopicName, opts.SubscriptionName, err)
	}

	return &ServiceBusConsumer{
		client:           client,
		receiver:         receiver,
		topicName:        opts.TopicName,
		subscriptionName: opts.SubscriptionName,
		maxRetries:       opts.MaxRetries,
		workerCount:      opts.WorkerCount,
		handler:          opts.Handler,
		publisher:        opts.Publisher,
	}, nil
}

// Start begins consuming messages
func (c *ServiceBusConsumer) Start(ctx context.Context) {
	log.Printf("🔄 Service Bus Consumer started for topic: %s, subscription: %s", c.topicName, c.subscriptionName)

	// Start worker pool
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx)
	}

	<-ctx.Done() // Wait for shutdown
	log.Println("🚦 Service Bus Consumer shutdown signal received")
}

// worker processes messages concurrently
func (c *ServiceBusConsumer) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("🚦 Shutdown signal received, stopping Service Bus worker...")
			return
		default:
			// Receive messages with timeout
			messages, err := c.receiver.ReceiveMessages(ctx, 1, nil)
			
			if err != nil {
				log.Printf("❌ Error receiving messages: %v", err)
				continue
			}

			// Process each message
			for _, message := range messages {
				c.processMessage(ctx, message)
			}
		}
	}
}

// processMessage handles a single message (retries, DLQ, etc.)
func (c *ServiceBusConsumer) processMessage(ctx context.Context, message *azservicebus.ReceivedMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Panic in Service Bus message handler: %v", r)
			// Abandon the message on panic
			if err := c.receiver.AbandonMessage(ctx, message, nil); err != nil {
				log.Printf("⚠️ Failed to abandon message after panic: %v", err)
			}
		}
	}()

	// Get retry count from message properties
	retryCount := c.getRetryCount(message)

	if retryCount >= c.maxRetries {
		log.Printf("❌ Max retries reached for message, moving to DLQ")
		// Dead letter the message
		reason := "MaxRetriesExceeded"
		description := fmt.Sprintf("Message failed after %d retries", retryCount)
		if err := c.receiver.DeadLetterMessage(ctx, message, &azservicebus.DeadLetterOptions{
			Reason:              &reason,
			ErrorDescription:    &description,
		}); err != nil {
			log.Printf("⚠️ Failed to dead letter message: %v", err)
		}
		return
	}

	// Process message
	if err := c.handler(ctx, message.Body); err != nil {
		log.Printf("❌ Handler failed, requeuing with delay (retry %d/%d)", retryCount+1, c.maxRetries)
		log.Printf("Error: %v", err)
		c.requeueWithDelay(ctx, message, retryCount)
		return
	}

	// Success - complete the message
	if err := c.receiver.CompleteMessage(ctx, message, nil); err != nil {
		log.Printf("⚠️ Failed to complete message: %v", err)
	}
}

// requeueWithDelay re-queues a message with exponential backoff
func (c *ServiceBusConsumer) requeueWithDelay(ctx context.Context, message *azservicebus.ReceivedMessage, retryCount int) {
	// Calculate delay with exponential backoff
	baseDelay := 1 << retryCount
	delay := time.Duration(baseDelay) * time.Second

	// Create new message with retry count
	newMessage := &azservicebus.Message{
		Body: message.Body,
		ApplicationProperties: make(map[string]interface{}),
	}

	// Copy existing properties
	for k, v := range message.ApplicationProperties {
		newMessage.ApplicationProperties[k] = v
	}

	// Update retry count
	newMessage.ApplicationProperties["x-retry-count"] = retryCount + 1

	// Schedule for later delivery
	scheduledTime := time.Now().Add(delay)
	newMessage.ScheduledEnqueueTime = &scheduledTime

	// Publish the new message using the publisher
	if err := c.publisher.PublishWithDelay(ctx, c.topicName, "", newMessage, delay); err != nil {
		log.Printf("❌ Failed to requeue message: %v", err)
		// Abandon the original message if requeue fails
		if abandonErr := c.receiver.AbandonMessage(ctx, message, nil); abandonErr != nil {
			log.Printf("⚠️ Failed to abandon message: %v", abandonErr)
		}
		return
	}

	// Complete the original message since we've successfully requeued
	if err := c.receiver.CompleteMessage(ctx, message, nil); err != nil {
		log.Printf("⚠️ Failed to complete original message: %v", err)
		return
	}

	log.Printf("↩️ Requeued message with delay (retry %d/%d)", retryCount+1, c.maxRetries)
}

// getRetryCount extracts retry count from message properties
func (c *ServiceBusConsumer) getRetryCount(message *azservicebus.ReceivedMessage) int {
	if retryCountRaw, exists := message.ApplicationProperties["x-retry-count"]; exists {
		switch v := retryCountRaw.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		default:
			return 0
		}
	}
	return 0
}

// Close closes the consumer and releases resources
func (c *ServiceBusConsumer) Close() error {
	if c.receiver != nil {
		if err := c.receiver.Close(context.Background()); err != nil {
			log.Printf("Error closing Service Bus receiver: %v", err)
		}
	}
	if c.client != nil {
		if err := c.client.Close(context.Background()); err != nil {
			log.Printf("Error closing Service Bus client: %v", err)
		}
	}
	return nil
}
