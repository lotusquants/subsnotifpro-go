# Deployment Guide

This guide covers different deployment scenarios for the subsnotifpro-go application, including both messaging and database configurations.

## Deployment Scenarios

### 1. Local Development (Container-based)

**Use Case**: Local development with Docker containers for both database and messaging.

**Configuration**:
```bash
# Messaging
MESSAGING_TYPE=rabbitmq

# Database
DB_DEPLOYMENT_MODE=container
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
```

**Deployment Steps**:
```bash
# 1. Start all services
docker-compose --profile postgres --profile rabbitmq up -d

# 2. Build and run
go build ./cmd/api
go build ./cmd/worker
./api &
./worker &
```

### 2. Hybrid Development (Container DB + Cloud Messaging)

**Use Case**: Local development with container database but cloud messaging for testing.

**Configuration**:
```bash
# Messaging
MESSAGING_TYPE=servicebus
SERVICEBUS_CONNECTION_STRING=your-servicebus-connection-string

# Database
DB_DEPLOYMENT_MODE=container
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
```

**Deployment Steps**:
```bash
# 1. Start database only
docker-compose --profile postgres up -d

# 2. Configure Service Bus connection
# 3. Build and run
go build ./cmd/api
go build ./cmd/worker
./api &
./worker &
```

### 3. Staging Environment (Managed Database + Cloud Messaging)

**Use Case**: Staging environment with managed database and cloud messaging.

**Configuration**:
```bash
# Messaging
MESSAGING_TYPE=servicebus
SERVICEBUS_USE_MANAGED_IDENTITY=true
SERVICEBUS_NAMESPACE=staging-namespace.servicebus.windows.net

# Database
DB_DEPLOYMENT_MODE=managed
DB_TYPE=postgres
DB_HOST=staging-db.postgres.database.azure.com
DB_PORT=5432
DB_NAME=subsnotifpro
DB_USER=staging_user@staging-db
DB_PASSWORD=staging_password
DB_SSLMODE=require
DB_AZURE_USE_MANAGED_IDENTITY=false
```

**Deployment Steps**:
```bash
# 1. Create Azure resources
az group create --name staging-rg --location eastus

# 2. Create managed database
az postgres server create \
    --name staging-db \
    --resource-group staging-rg \
    --location eastus \
    --admin-user staging_user \
    --admin-password staging_password \
    --sku-name GP_Gen5_2

# 3. Create Service Bus namespace
az servicebus namespace create \
    --name staging-namespace \
    --resource-group staging-rg \
    --location eastus

# 4. Deploy application
# (depends on your deployment target: Container Apps, App Service, etc.)
```

### 4. Production Environment (Fully Managed)

**Use Case**: Production environment with managed database, cloud messaging, and managed identity.

**Configuration**:
```bash
# Messaging
MESSAGING_TYPE=servicebus
SERVICEBUS_USE_MANAGED_IDENTITY=true
SERVICEBUS_NAMESPACE=prod-namespace.servicebus.windows.net

# Database
DB_DEPLOYMENT_MODE=managed
DB_TYPE=postgres
DB_HOST=prod-db.postgres.database.azure.com
DB_PORT=5432
DB_NAME=subsnotifpro
DB_USER=prod_user@prod-db
DB_SSLMODE=require
DB_AZURE_USE_MANAGED_IDENTITY=true
DB_AZURE_RESOURCE_ID=/subscriptions/your-sub/resourceGroups/prod-rg/providers/Microsoft.DBforPostgreSQL/servers/prod-db
```

**Deployment Steps**:
```bash
# 1. Create Azure resources with enhanced security
az group create --name prod-rg --location eastus

# 2. Create managed database with advanced security
az postgres server create \
    --name prod-db \
    --resource-group prod-rg \
    --location eastus \
    --admin-user prod_user \
    --admin-password prod_password \
    --sku-name GP_Gen5_4 \
    --ssl-enforcement Enabled \
    --minimal-tls-version TLS1_2

# 3. Create Service Bus namespace
az servicebus namespace create \
    --name prod-namespace \
    --resource-group prod-rg \
    --location eastus \
    --sku Premium

# 4. Configure managed identity and RBAC
# 5. Deploy application with managed identity
```

## Docker Compose Profiles

Use profiles to control which services are started:

```bash
# Database only
docker-compose --profile postgres up -d

# Messaging only
docker-compose --profile rabbitmq up -d

# Both database and messaging
docker-compose --profile postgres --profile rabbitmq up -d

# Application services only (for managed database/messaging)
docker-compose up -d
```

## Environment Configuration Matrix

| Scenario | Database | Messaging | Docker Services |
|----------|----------|-----------|----------------|
| Local Dev | Container | Container | postgres, rabbitmq |
| Hybrid Dev | Container | Managed | postgres |
| Staging | Managed | Managed | none |
| Production | Managed | Managed | none |

## Security Configuration

### Local Development
- SSL disabled for simplicity
- Default credentials
- No network restrictions

### Staging/Production
- SSL/TLS enforced
- Strong passwords or managed identity
- Network security groups and firewall rules
- Private endpoints (recommended for production)

## Monitoring and Observability

### Container-based
- Docker logs: `docker-compose logs`
- Local monitoring tools

### Managed Services
- Azure Monitor and Application Insights
- Database query performance insights
- Service Bus metrics and alerts

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Build
      run: |
        go build ./cmd/api
        go build ./cmd/worker
    
    - name: Deploy to Azure
      env:
        MESSAGING_TYPE: servicebus
        DB_DEPLOYMENT_MODE: managed
        AZURE_SERVICEBUS_CONNECTION_STRING: ${{ secrets.AZURE_SERVICEBUS_CONNECTION_STRING }}
        DB_HOST: ${{ secrets.DB_HOST }}
        DB_USER: ${{ secrets.DB_USER }}
        DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
        DB_NAME: ${{ secrets.DB_NAME }}
        DB_SSL_MODE: require
      run: |
        # Deploy to Azure Container Apps or App Service
        # Example using Azure CLI
        az containerapp update \
          --name subsnotifpro-api \
          --resource-group your-resource-group \
          --image your-registry/subsnotifpro-api:latest \
          --set-env-vars \
            MESSAGING_TYPE=servicebus \
            DB_DEPLOYMENT_MODE=managed \
            AZURE_SERVICEBUS_CONNECTION_STRING="$AZURE_SERVICEBUS_CONNECTION_STRING" \
            DB_HOST="$DB_HOST" \
            DB_USER="$DB_USER" \
            DB_PASSWORD="$DB_PASSWORD" \
            DB_NAME="$DB_NAME" \
            DB_SSL_MODE=require
```

## Troubleshooting

### Common Issues

1. **Database Connection Failures**
   - Check firewall rules
   - Verify SSL configuration
   - Confirm credentials

2. **Messaging Connection Failures**
   - Check Service Bus namespace
   - Verify connection string or managed identity
   - Confirm topic/queue existence

3. **Performance Issues**
   - Adjust connection pool settings
   - Monitor database query performance
   - Check messaging throughput limits

### Diagnostic Commands

```bash
# Test database connectivity
psql -h your-host -p 5432 -U your-user -d your-database

# Test database with SSL
psql "host=your-host port=5432 user=your-user dbname=your-database sslmode=require"

# Test RabbitMQ connectivity
curl -u admin:admin123 http://localhost:15672/api/overview

# Test Azure Service Bus connectivity (using Azure CLI)
az servicebus namespace show --name your-namespace --resource-group your-rg

# Test application endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/status

# Check Docker service logs
docker-compose logs postgres
docker-compose logs rabbitmq

# Monitor connection pools and performance
# (Application logs will show connection pool configuration and status)
```

### Health Check Endpoints

Add these to your application for better monitoring:

```go
// Add to your routes
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now().UTC(),
        "version": "1.0.0",
    })
})

router.GET("/api/v1/status", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "database": "connected",
        "messaging": "connected",
        "services": "operational",
    })
})
```

## Scaling Considerations

### Database Scaling
- **Container**: Limited to single instance
- **Managed**: Supports read replicas and scaling
- **External**: Depends on provider capabilities

### Messaging Scaling
- **RabbitMQ**: Cluster setup for high availability
- **Service Bus**: Built-in scaling and partitioning

## Cost Optimization

### Development
- Use container-based services
- Shut down services when not needed
- Use smaller SKUs for managed services

### Production
- Right-size managed services
- Use reserved instances for predictable workloads
- Monitor and optimize based on usage patterns

## Best Practices

1. **Environment Isolation**: Use separate resources for dev/staging/prod
2. **Security**: Enable SSL, use managed identity, implement network security
3. **Monitoring**: Set up comprehensive monitoring and alerting
4. **Backup**: Implement automated backup strategies
5. **Documentation**: Keep deployment documentation updated
6. **Testing**: Test deployment procedures in staging environment
7. **Rollback**: Have rollback procedures ready for production deployments
