package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ServiceBusPublisher implements MessagePublisher for Azure Service Bus
type ServiceBusPublisher struct {
	client   *azservicebus.Client
	sender   *azservicebus.Sender
	topicName string
}

// ServiceBusPublisherOptions contains configuration for ServiceBus publisher
type ServiceBusPublisherOptions struct {
	Namespace         string
	ConnectionString  string
	UseManagedIdentity bool
	TopicName         string
}

// NewServiceBusPublisher creates a new Azure Service Bus publisher
func NewServiceBusPublisher(opts ServiceBusPublisherOptions) (*ServiceBusPublisher, error) {
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

	// Create sender for the topic
	sender, err := client.NewSender(opts.TopicName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender for topic %s: %w", opts.TopicName, err)
	}

	return &ServiceBusPublisher{
		client:   client,
		sender:   sender,
		topicName: opts.TopicName,
	}, nil
}

// PublishToQueue publishes directly to a queue (maps to topic for Service Bus)
func (p *ServiceBusPublisher) PublishToQueue(ctx context.Context, queueName string, event interface{}) error {
	return p.PublishToExchange(ctx, queueName, "", event)
}

// PublishToExchange publishes to a topic (Service Bus equivalent of exchange)
func (p *ServiceBusPublisher) PublishToExchange(ctx context.Context, topicName, routingKey string, event interface{}) error {
	// Marshal the event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create the message
	message := &azservicebus.Message{
		Body: body,
		ApplicationProperties: map[string]interface{}{
			"ContentType": "application/json",
			"Timestamp":   time.Now().Unix(),
		},
	}

	// Add routing key as a property if provided
	if routingKey != "" {
		message.ApplicationProperties["RoutingKey"] = routingKey
	}

	// Send the message
	if err := p.sender.SendMessage(ctx, message, nil); err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", topicName, err)
	}

	log.Printf("Published message to Service Bus topic: %s", topicName)
	return nil
}

// PublishWithDelay publishes with a delay (using ScheduledEnqueueTime)
func (p *ServiceBusPublisher) PublishWithDelay(ctx context.Context, topicName, routingKey string, event interface{}, delay time.Duration) error {
	// Marshal the event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create the message with scheduled enqueue time
	scheduledTime := time.Now().Add(delay)
	message := &azservicebus.Message{
		Body: body,
		ApplicationProperties: map[string]interface{}{
			"ContentType": "application/json",
			"Timestamp":   time.Now().Unix(),
		},
		ScheduledEnqueueTime: &scheduledTime,
	}

	// Add routing key as a property if provided
	if routingKey != "" {
		message.ApplicationProperties["RoutingKey"] = routingKey
	}

	// Send the message
	if err := p.sender.SendMessage(ctx, message, nil); err != nil {
		return fmt.Errorf("failed to send delayed message to topic %s: %w", topicName, err)
	}

	log.Printf("Published delayed message to Service Bus topic: %s (delay: %v)", topicName, delay)
	return nil
}

// Close closes the publisher and releases resources
func (p *ServiceBusPublisher) Close() error {
	if p.sender != nil {
		if err := p.sender.Close(context.Background()); err != nil {
			log.Printf("Error closing Service Bus sender: %v", err)
		}
	}
	if p.client != nil {
		if err := p.client.Close(context.Background()); err != nil {
			log.Printf("Error closing Service Bus client: %v", err)
		}
	}
	return nil
}
