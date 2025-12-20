# Messaging Backend Configuration

This project supports two messaging backends: **RabbitMQ** and **Azure Service Bus**. You can switch between them using the `MESSAGING_TYPE` environment variable.

## Configuration

### Environment Variables

The messaging backend is selected using the `MESSAGING_TYPE` environment variable:

```bash
# For RabbitMQ
MESSAGING_TYPE=rabbitmq

# For Azure Service Bus
MESSAGING_TYPE=servicebus
```

### RabbitMQ Configuration

When using RabbitMQ (`MESSAGING_TYPE=rabbitmq`), configure the following environment variables:

```bash
# RabbitMQ Connection
RABBITMQ_USERNAME=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_VHOST=/
RABBITMQ_CONNECTION_NAME=subsnotifpro-connection

# RabbitMQ Behavior
RABBITMQ_MAX_RETRIES=3
RABBITMQ_RETRY_DELAY=5s
RABBITMQ_WORKER_COUNT=5

# Queue Configuration (see .env.example for complete list)
RABBITMQ_RTDN_EXCHANGE=google-play-rtdn
RABBITMQ_RTDN_QUEUE=google-play-rtdn-queue
# ... additional queue configurations
```

### Azure Service Bus Configuration

When using Azure Service Bus (`MESSAGING_TYPE=servicebus`), configure the following environment variables:

```bash
# Service Bus Connection (choose one method)
# Method 1: Connection String
SERVICEBUS_CONNECTION_STRING=Endpoint=sb://your-namespace.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=your-key

# Method 2: Managed Identity (recommended for production)
SERVICEBUS_NAMESPACE=your-namespace.servicebus.windows.net
SERVICEBUS_USE_MANAGED_IDENTITY=true

# Service Bus Behavior
SERVICEBUS_MAX_RETRIES=3
SERVICEBUS_WORKER_COUNT=5

# Topic Configuration (see .env.example for complete list)
SERVICEBUS_RTDN_TOPIC=google-play-rtdn
SERVICEBUS_RTDN_SUBSCRIPTION=google-play-rtdn-subscription
# ... additional topic configurations
```

## Deployment Options

### Option 1: RabbitMQ with Docker Compose

1. Set `MESSAGING_TYPE=rabbitmq` in your environment
2. Use the provided Docker Compose file:

```bash
# Start RabbitMQ
docker-compose --profile rabbitmq up -d

# Run the application
go run ./cmd/api
go run ./cmd/worker
```

### Option 2: Azure Service Bus

1. Set `MESSAGING_TYPE=servicebus` in your environment
2. Create Azure Service Bus namespace and topics:

```bash
# Create Service Bus namespace
az servicebus namespace create --name your-namespace --resource-group your-rg

# Create topics
az servicebus topic create --namespace-name your-namespace --name google-play-rtdn --resource-group your-rg
az servicebus topic create --namespace-name your-namespace --name app-store-notifications --resource-group your-rg
az servicebus topic create --namespace-name your-namespace --name unified-subscriptions --resource-group your-rg

# Create subscriptions
az servicebus topic subscription create --namespace-name your-namespace --topic-name google-play-rtdn --name google-play-rtdn-subscription --resource-group your-rg
az servicebus topic subscription create --namespace-name your-namespace --topic-name app-store-notifications --name app-store-notifications-subscription --resource-group your-rg
az servicebus topic subscription create --namespace-name your-namespace --topic-name unified-subscriptions --name unified-subscriptions-subscription --resource-group your-rg
```

3. Configure authentication:

**Using Connection String:**
```bash
# Get connection string
az servicebus namespace authorization-rule keys list --namespace-name your-namespace --name RootManageSharedAccessKey --resource-group your-rg --query primaryConnectionString -o tsv
```

**Using Managed Identity (recommended):**
```bash
# Assign Service Bus Data Owner role to your managed identity
az role assignment create --assignee your-managed-identity --role "Azure Service Bus Data Owner" --scope /subscriptions/your-subscription/resourceGroups/your-rg/providers/Microsoft.ServiceBus/namespaces/your-namespace
```

## Docker Compose Profiles

The Docker Compose file includes profiles for conditional service deployment:

```bash
# Start with RabbitMQ
docker-compose --profile rabbitmq up -d

# Start without RabbitMQ (for Service Bus)
docker-compose up -d
```

## Development

### Running Locally

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Edit `.env` and set your preferred messaging backend

3. Start the services:
```bash
# Build the application
go build ./cmd/api
go build ./cmd/worker

# Run the services
./api &
./worker &
```

### Testing Different Backends

You can easily test both backends by changing the `MESSAGING_TYPE` environment variable:

```bash
# Test with RabbitMQ
MESSAGING_TYPE=rabbitmq go run ./cmd/api

# Test with Azure Service Bus
MESSAGING_TYPE=servicebus go run ./cmd/api
```

## Architecture

The messaging abstraction is built around these key components:

- **MessagePublisher**: Interface for publishing messages
- **MessageConsumer**: Interface for consuming messages
- **MessagingFactory**: Factory for creating publishers and consumers
- **Configuration**: Unified configuration for both backends

The factory pattern allows seamless switching between RabbitMQ and Azure Service Bus without changing application code.

Similarly, the database layer supports multiple deployment modes (container, managed, external) with the same application code. See [DATABASE.md](DATABASE.md) for detailed database configuration.

## Monitoring and Troubleshooting

### RabbitMQ

- Management UI: http://localhost:15672 (guest/guest)
- Logs: `docker-compose logs rabbitmq`

### Azure Service Bus

- Azure Portal: Monitor metrics and logs
- Azure CLI: `az servicebus topic show --namespace-name your-namespace --name topic-name`

## Best Practices

1. **Use Managed Identity** for Azure Service Bus in production
2. **Configure retry policies** appropriate for your use case
3. **Monitor message processing** and dead letter queues
4. **Use connection pooling** for high-throughput scenarios
5. **Implement proper error handling** and logging
