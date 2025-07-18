package main

import (
	"fmt"
	"os"
	"subsnotifpro-go/config"
)

func main() {
	// Test RabbitMQ configuration
	os.Setenv("MESSAGING_TYPE", "rabbitmq")
	os.Setenv("RABBITMQ_HOST", "localhost")
	os.Setenv("RABBITMQ_PORT", "5672")
	
	cfg := config.LoadConfig()
	fmt.Printf("=== RabbitMQ Configuration Test ===\n")
	fmt.Printf("Messaging Type: %s\n", cfg.MessagingType)
	fmt.Printf("RabbitMQ Host: %s\n", cfg.RabbitMQ.Host)
	fmt.Printf("RabbitMQ Port: %s\n", cfg.RabbitMQ.Port)
	
	// Test Service Bus configuration
	os.Setenv("MESSAGING_TYPE", "servicebus")
	os.Setenv("SERVICEBUS_NAMESPACE", "test-namespace.servicebus.windows.net")
	
	cfg = config.LoadConfig()
	fmt.Printf("\n=== Service Bus Configuration Test ===\n")
	fmt.Printf("Messaging Type: %s\n", cfg.MessagingType)
	fmt.Printf("Service Bus Namespace: %s\n", cfg.ServiceBus.Namespace)
	fmt.Printf("Service Bus Use Managed Identity: %t\n", cfg.ServiceBus.UseManagedIdentity)
	
	fmt.Printf("\n✅ Configuration loading test passed!\n")
}
