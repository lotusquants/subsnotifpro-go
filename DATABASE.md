# Database Configuration Guide

This project supports multiple database deployment modes: **Container**, **Managed**, and **External**. You can switch between them using the `DB_DEPLOYMENT_MODE` environment variable.

## Configuration Overview

### Environment Variables

The database deployment mode is selected using the `DB_DEPLOYMENT_MODE` environment variable:

```bash
# For container-based PostgreSQL (Docker)
DB_DEPLOYMENT_MODE=container

# For managed databases (Azure Database for PostgreSQL)
DB_DEPLOYMENT_MODE=managed

# For external databases (hosted elsewhere)
DB_DEPLOYMENT_MODE=external
```

## Deployment Modes

### 1. Container Deployment (`container`)

**Use Case**: Local development, testing, and environments where you want to run PostgreSQL in Docker containers.

**Configuration**:
```bash
# Database deployment
DB_DEPLOYMENT_MODE=container
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# Connection pool settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=15m
```

**Docker Setup**:
```bash
# Start PostgreSQL container
docker-compose --profile postgres up -d

# Verify connection
docker exec -it subsnotifpro-postgres psql -U postgres -d subsnotifpro_db
```

### 2. Managed Deployment (`managed`)

**Use Case**: Production and staging environments using Azure Database for PostgreSQL or similar managed services.

**Configuration**:
```bash
# Database deployment
DB_DEPLOYMENT_MODE=managed
DB_TYPE=postgres
DB_HOST=your-server.postgres.database.azure.com
DB_PORT=5432
DB_NAME=subsnotifpro
DB_USER=your-admin@your-server
DB_PASSWORD=your-secure-password
DB_SSLMODE=require

# Azure-specific settings
DB_AZURE_USE_MANAGED_IDENTITY=false
DB_AZURE_SERVER_NAME=your-server.postgres.database.azure.com
DB_AZURE_RESOURCE_ID=/subscriptions/your-subscription/resourceGroups/your-rg/providers/Microsoft.DBforPostgreSQL/servers/your-server
DB_AZURE_TENANT_ID=your-tenant-id
DB_AZURE_SSL_MODE=require
DB_AZURE_CONNECT_TIMEOUT=30s
DB_AZURE_READ_TIMEOUT=30s
DB_AZURE_WRITE_TIMEOUT=30s

# Connection pool settings (optimized for managed services)
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=30m
```

**Azure Setup**:
```bash
# Create Azure Database for PostgreSQL
az postgres server create \
    --name your-server \
    --resource-group your-rg \
    --location eastus \
    --admin-user your-admin \
    --admin-password your-secure-password \
    --sku-name GP_Gen5_2 \
    --ssl-enforcement Enabled \
    --minimal-tls-version TLS1_2

# Create database
az postgres db create \
    --name subsnotifpro \
    --server-name your-server \
    --resource-group your-rg

# Configure firewall (for development - use private endpoints in production)
az postgres server firewall-rule create \
    --name AllowAzureIPs \
    --server your-server \
    --resource-group your-rg \
    --start-ip-address 0.0.0.0 \
    --end-ip-address 0.0.0.0
```

### 3. External Deployment (`external`)

**Use Case**: Connecting to databases hosted by other providers or self-managed PostgreSQL instances.

**Configuration**:
```bash
# Database deployment
DB_DEPLOYMENT_MODE=external
DB_TYPE=postgres
DB_HOST=external-db.example.com
DB_PORT=5432
DB_NAME=production_db
DB_USER=app_user
DB_PASSWORD=secure_password
DB_SSLMODE=require

# Connection pool settings
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=15m
```

```bash
# For container-based PostgreSQL (Docker)
DB_DEPLOYMENT_MODE=container

# For managed database (Azure Database for PostgreSQL)
DB_DEPLOYMENT_MODE=managed

# For external database (self-hosted or third-party)
DB_DEPLOYMENT_MODE=external
```

### Container Deployment (Docker)

When using container deployment (`DB_DEPLOYMENT_MODE=container`), configure the following environment variables:

```bash
# Database Configuration
DB_DEPLOYMENT_MODE=container
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# Connection Pool Settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=15m

# Additional Settings
DB_QUERY_LOGGING=false
DB_SSL_REQUIRED=false
```

### Managed Database (Azure Database for PostgreSQL)

When using managed database (`DB_DEPLOYMENT_MODE=managed`), configure the following environment variables:

```bash
# Database Configuration
DB_DEPLOYMENT_MODE=managed
DB_TYPE=postgres
DB_HOST=your-server.postgres.database.azure.com
DB_PORT=5432
DB_NAME=subsnotifpro
DB_USER=your-username@your-server
DB_PASSWORD=your-password
DB_SSLMODE=require

# Azure-specific Configuration
DB_AZURE_USE_MANAGED_IDENTITY=false
DB_AZURE_SERVER_NAME=your-server.postgres.database.azure.com
DB_AZURE_RESOURCE_ID=/subscriptions/your-subscription/resourceGroups/your-rg/providers/Microsoft.DBforPostgreSQL/servers/your-server
DB_AZURE_TENANT_ID=your-tenant-id
DB_AZURE_SSL_MODE=require
DB_AZURE_SSL_ROOT_CERT=/path/to/ssl/root/cert.pem

# Connection Pool Settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=15m

# Additional Settings
DB_QUERY_LOGGING=false
DB_SSL_REQUIRED=true
```

### External Database

When using external database (`DB_DEPLOYMENT_MODE=external`), configure the following environment variables:

```bash
# Database Configuration
DB_DEPLOYMENT_MODE=external
DB_TYPE=postgres
DB_HOST=your-external-db-host
DB_PORT=5432
DB_NAME=your-database
DB_USER=your-username
DB_PASSWORD=your-password
DB_SSLMODE=disable

# Connection Pool Settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=15m

# Additional Settings
DB_QUERY_LOGGING=false
DB_SSL_REQUIRED=false
```

## Deployment Options

### Option 1: Container Deployment with Docker Compose

1. Set `DB_DEPLOYMENT_MODE=container` in your environment
2. Use the provided Docker Compose file:

```bash
# Start PostgreSQL container
docker-compose --profile postgres up -d

# Or start with both database and messaging
docker-compose --profile postgres --profile rabbitmq up -d

# Run the application
go run ./cmd/api
go run ./cmd/worker
```

### Option 2: Managed Database (Azure)

1. Set `DB_DEPLOYMENT_MODE=managed` in your environment
2. Create Azure Database for PostgreSQL:

```bash
# Create Azure Database for PostgreSQL
az postgres server create \
    --name your-server \
    --resource-group your-rg \
    --location eastus \
    --admin-user your-username \
    --admin-password your-password \
    --sku-name GP_Gen5_2 \
    --version 13

# Configure firewall (allow Azure services)
az postgres server firewall-rule create \
    --resource-group your-rg \
    --server your-server \
    --name AllowAzureServices \
    --start-ip-address 0.0.0.0 \
    --end-ip-address 0.0.0.0

# Create database
az postgres db create \
    --resource-group your-rg \
    --server-name your-server \
    --name subsnotifpro
```

3. Configure authentication:

**Using Connection String:**
```bash
# Standard authentication with username/password
DB_AZURE_USE_MANAGED_IDENTITY=false
DB_USER=your-username@your-server
DB_PASSWORD=your-password
```

**Using Managed Identity (recommended):**
```bash
# Enable managed identity authentication
DB_AZURE_USE_MANAGED_IDENTITY=true
DB_AZURE_TENANT_ID=your-tenant-id
DB_AZURE_RESOURCE_ID=/subscriptions/your-subscription/resourceGroups/your-rg/providers/Microsoft.DBforPostgreSQL/servers/your-server
```

### Option 3: External Database

1. Set `DB_DEPLOYMENT_MODE=external` in your environment
2. Configure connection details for your external database
3. Ensure network connectivity and firewall rules are properly configured

## Database Types

### PostgreSQL (Recommended)

```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
```

### MySQL

```bash
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=subsnotifpro_db
DB_USER=root
DB_PASSWORD=password
```

### SQLite (Development Only)

```bash
DB_TYPE=sqlite
DB_NAME=subsnotifpro.db
```

## Docker Compose Profiles

The Docker Compose file includes profiles for conditional database deployment:

```bash
# Start with PostgreSQL
docker-compose --profile postgres up -d

# Start with PostgreSQL and RabbitMQ
docker-compose --profile postgres --profile rabbitmq up -d

# Start without database (for managed/external)
docker-compose up -d
```

## Connection Pool Configuration

The application supports configurable connection pooling:

```bash
# Maximum number of open connections
DB_MAX_OPEN_CONNS=25

# Maximum number of idle connections
DB_MAX_IDLE_CONNS=5

# Maximum lifetime of a connection
DB_CONN_MAX_LIFETIME=15m
```

## Security Considerations

### SSL/TLS Configuration

- **Container**: SSL typically disabled for simplicity
- **Managed**: SSL required and enforced
- **External**: SSL recommended for production

### Authentication

- **Container**: Username/password authentication
- **Managed**: Support for managed identity and connection strings
- **External**: Username/password or certificate-based authentication

### Network Security

- **Container**: Isolated Docker network
- **Managed**: Azure private endpoints and firewall rules
- **External**: Network-level security controls

## Performance Optimization

### Connection Pooling

```bash
# Optimize for high-throughput applications
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=30m
```

### Query Logging

```bash
# Enable for debugging (disable in production)
DB_QUERY_LOGGING=true
```

## Monitoring and Troubleshooting

### Container Deployment

- Database logs: `docker-compose logs postgres`
- Database shell: `docker exec -it subsnotifpro-postgres psql -U postgres -d subsnotifpro_db`

### Managed Database

- Azure Portal: Monitor metrics and query performance
- Azure CLI: `az postgres server show --resource-group your-rg --name your-server`

### External Database

- Use database-specific monitoring tools
- Check network connectivity and firewall rules

## Migration and Backup

### Automatic Migrations

The application automatically creates and updates database schema on startup:

```go
database.AutoMigrateTables(db, cfg)
```

### Manual Backup

```bash
# PostgreSQL backup
pg_dump -h localhost -U postgres subsnotifpro_db > backup.sql

# PostgreSQL restore
psql -h localhost -U postgres subsnotifpro_db < backup.sql
```

## Development Workflow

### Local Development

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Choose your database deployment mode:
```bash
# For Docker
DB_DEPLOYMENT_MODE=container

# For managed database
DB_DEPLOYMENT_MODE=managed
```

3. Start the database:
```bash
# Container
docker-compose --profile postgres up -d

# Managed/External
# Database should already be running
```

4. Run the application:
```bash
go run ./cmd/api
go run ./cmd/worker
```

### Testing Different Deployments

```bash
# Test with container
DB_DEPLOYMENT_MODE=container go run ./cmd/api

# Test with managed database
DB_DEPLOYMENT_MODE=managed go run ./cmd/api
```

## Best Practices

1. **Use managed databases** for production deployments
2. **Enable SSL/TLS** for all external connections
3. **Configure connection pooling** based on your workload
4. **Monitor database performance** and query patterns
5. **Implement proper backup strategies**
6. **Use environment-specific configurations**
7. **Enable query logging** only for debugging
8. **Follow security best practices** for authentication and authorization

## Architecture

The database abstraction supports multiple deployment modes:

```
┌─────────────────────────────────────────────┐
│                Application                  │
├─────────────────────────────────────────────┤
│           Database Configuration            │
├─────────────┬──────────────┬────────────────┤
│  Container  │   Managed    │   External     │
│  PostgreSQL │   Database   │   Database     │
│  (Docker)   │   (Azure)    │   (Custom)     │
└─────────────┴──────────────┴────────────────┘
```

This provides flexibility for different deployment scenarios while maintaining consistent application code.
