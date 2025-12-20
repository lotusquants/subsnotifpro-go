# 🚀 SubsNotifPro Go - Enhanced Edition

## 🎯 Overview

SubsNotifPro Go is now a **production-ready, enterprise-grade** application with flexible backend support for both messaging and database layers. This enhanced version includes:

### ✨ New Features

- **🏥 Advanced Health Checks** - Comprehensive health monitoring for all components
- **🧪 Automated Testing Suite** - Connectivity tests, integration tests, and health checks
- **📊 Monitoring & Alerting** - Real-time monitoring with automated alerts
- **🔧 Database Management** - Automated backup, restore, and migration scripts
- **🐳 Container Support** - Production-ready Docker configuration
- **📈 Performance Metrics** - System metrics and performance monitoring

## 🏗️ Architecture

### Backend Flexibility

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│  🔄 Messaging Layer    │  🗄️ Database Layer               │
│  ├─ RabbitMQ (Self)    │  ├─ Container (Docker)            │
│  └─ Service Bus (Azure)│  ├─ Managed (Azure DB)            │
│                        │  └─ External (Custom)             │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure                           │
│  🐳 Docker Compose  │  ☁️ Azure Services  │  🔧 Scripts   │
└─────────────────────────────────────────────────────────────┘
```

### Health & Monitoring

```
┌─────────────────────────────────────────────────────────────┐
│                    Health Monitoring                        │
├─────────────────────────────────────────────────────────────┤
│  🏥 Health Checks      │  📊 Metrics Collection             │
│  ├─ Database Health    │  ├─ CPU, Memory, Disk              │
│  ├─ Messaging Health   │  ├─ Connection Pools               │
│  ├─ API Health         │  └─ Response Times                 │
│  └─ System Health      │                                    │
├─────────────────────────────────────────────────────────────┤
│  🚨 Alerting           │  📋 Reporting                      │
│  ├─ Webhook Alerts     │  ├─ Health Reports                 │
│  ├─ Email Notifications│  ├─ Test Reports                  │
│  └─ Log Aggregation    │  └─ Monitoring Reports             │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### 1. Local Development Setup

```bash
# Clone and setup
git clone <repository>
cd subsnotifpro-go

# Configuration
cp .env.example .env
# Edit .env with your settings

# Start all services
docker-compose --profile postgres --profile rabbitmq up -d

# Run tests
./scripts/test_runner.sh all

# Start application
go run ./cmd/api &
go run ./cmd/worker &
```

### 2. Production Deployment

```bash
# Build production image
docker build -t subsnotifpro-go:latest .

# Deploy with managed services
export DB_DEPLOYMENT_MODE=managed
export MESSAGING_TYPE=servicebus
# ... set other production variables

# Deploy to Azure Container Apps
az containerapp create \
  --name subsnotifpro-api \
  --resource-group myResourceGroup \
  --image subsnotifpro-go:latest \
  --environment myEnvironment
```

## 🏥 Health & Monitoring

### Health Check Endpoints

- **`/health`** - Comprehensive health check with component status
- **`/health/ready`** - Readiness probe for orchestrators
- **`/health/live`** - Liveness probe for orchestrators
- **`/api/health`** - Legacy health check (backwards compatibility)

### Example Health Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "database": {
      "status": "healthy",
      "message": "Database is healthy",
      "last_checked": "2024-01-15T10:30:00Z",
      "response_time_ms": 12
    },
    "messaging": {
      "status": "healthy",
      "message": "RabbitMQ is healthy",
      "last_checked": "2024-01-15T10:30:00Z",
      "response_time_ms": 8
    }
  },
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

## 🧪 Testing

### Test Suite Overview

The application includes comprehensive testing capabilities:

```bash
# Run all tests
./scripts/test_runner.sh all

# Run specific test categories
./scripts/test_runner.sh unit           # Unit tests
./scripts/test_runner.sh connectivity   # Database/messaging connectivity
./scripts/test_runner.sh integration    # Integration tests
./scripts/test_runner.sh health         # Health check endpoint tests
./scripts/test_runner.sh build          # Build verification tests
./scripts/test_runner.sh config         # Configuration validation tests
```

### Connectivity Tests

The connectivity test suite validates:

- **Database connectivity** across all deployment modes
- **RabbitMQ connectivity** with channel and consumer tests
- **Azure Service Bus connectivity** with sender/receiver tests
- **Connection pool management** and concurrent access
- **Error handling** and retry mechanisms

### Health Check Tests

Automated tests verify:

- All health endpoints respond correctly
- Component health checks work properly
- Status codes match expected values
- Response formats are correct
- Readiness and liveness probes function

## 📊 Database Management

### Automated Database Operations

```bash
# Initialize database
./scripts/db_manager.sh init

# Create backup
./scripts/db_manager.sh backup

# Restore from backup
./scripts/db_manager.sh restore backup_20240115_103000.sql.gz

# Check database status
./scripts/db_manager.sh status

# Run migrations
./scripts/db_manager.sh migrate
```

### Migration Support

The database manager supports:

- **Automated migrations** with golang-migrate
- **Backup and restore** with compression
- **Database reset** for development
- **Connection testing** across all modes
- **Status monitoring** with detailed metrics

## 🔍 Monitoring & Alerting

### Real-time Monitoring

```bash
# Start continuous monitoring
./scripts/monitor.sh monitor

# Run single health check
./scripts/monitor.sh check

# Generate monitoring report
./scripts/monitor.sh report

# Setup as system service
sudo ./scripts/monitor.sh setup
```

### Monitoring Features

- **Continuous health monitoring** with configurable intervals
- **Automated alerting** via webhooks and email
- **System metrics collection** (CPU, memory, disk, load)
- **Container status monitoring** for Docker deployments
- **Process monitoring** for application components
- **Report generation** with comprehensive status information

### Alert Configuration

```bash
# Set environment variables for alerting
export WEBHOOK_URL="https://hooks.slack.com/services/..."
export ALERT_EMAIL="admin@company.com"
export CHECK_INTERVAL=60
export MAX_FAILURES=3
```

## 🐳 Container Support

### Docker Deployment

```bash
# Build image
docker build -t subsnotifpro-go:latest .

# Run with container dependencies
docker-compose --profile postgres --profile rabbitmq up -d

# Run API container
docker run -d \
  --name subsnotifpro-api \
  --network subsnotifpro-network \
  -p 8080:8080 \
  -e DB_DEPLOYMENT_MODE=container \
  -e MESSAGING_TYPE=rabbitmq \
  subsnotifpro-go:latest
```

### Container Features

- **Multi-stage build** for optimized image size
- **Non-root user** for security
- **Health checks** built into container
- **Graceful shutdown** handling
- **Configuration via environment variables**

## 📈 Performance & Scaling

### Performance Monitoring

The application includes built-in performance monitoring:

- **Connection pool metrics** for database connections
- **Response time tracking** for health checks
- **System resource monitoring** (CPU, memory, disk)
- **Load average tracking** for system health
- **Request/response metrics** for API endpoints

### Scaling Considerations

- **Connection pooling** optimized for concurrent access
- **Messaging queue management** with proper acknowledgments
- **Database connection limits** with pool configuration
- **Health check optimization** to avoid resource exhaustion
- **Graceful shutdown** for zero-downtime deployments

## 🔧 Configuration

### Environment Variables

#### Core Configuration
```bash
# Backend Selection
MESSAGING_TYPE=rabbitmq|servicebus
DB_DEPLOYMENT_MODE=container|managed|external

# Server Configuration
SERVER_PORT=8080
API_URL=http://localhost:8080
```

#### Database Configuration
```bash
# All Modes
DB_HOST=localhost
DB_PORT=5432
DB_NAME=subsnotifpro_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# Managed Mode Additional
DB_AZURE_USE_MANAGED_IDENTITY=true
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
```

#### Messaging Configuration
```bash
# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_MANAGEMENT_PORT=15672

# Azure Service Bus
SERVICEBUS_CONNECTION_STRING=Endpoint=sb://...
SERVICEBUS_NAMESPACE=your-namespace.servicebus.windows.net
SERVICEBUS_USE_MANAGED_IDENTITY=true
```

#### Monitoring Configuration
```bash
# Health Check Configuration
HEALTH_ENDPOINT=/health
READINESS_ENDPOINT=/health/ready
LIVENESS_ENDPOINT=/health/live

# Monitoring Configuration
CHECK_INTERVAL=60
TIMEOUT=10
MAX_FAILURES=3
WEBHOOK_URL=https://hooks.slack.com/services/...
ALERT_EMAIL=admin@company.com
```

## 🛠️ Development

### Project Structure

```
subsnotifpro-go/
├── cmd/
│   ├── api/main.go          # API server entrypoint
│   └── worker/main.go       # Worker service entrypoint
├── internal/
│   ├── health/              # Health check system
│   │   └── health.go
│   ├── tests/               # Test suites
│   │   └── connectivity_test.go
│   └── ...
├── scripts/
│   ├── db_manager.sh        # Database management
│   ├── monitor.sh           # Monitoring and alerting
│   └── test_runner.sh       # Test execution
├── Dockerfile               # Production container
├── docker-compose.yml       # Development environment
└── README.md               # This file
```

### Development Workflow

1. **Setup Environment**
   ```bash
   cp .env.example .env
   # Edit configuration
   ```

2. **Start Dependencies**
   ```bash
   docker-compose --profile postgres --profile rabbitmq up -d
   ```

3. **Run Tests**
   ```bash
   ./scripts/test_runner.sh all
   ```

4. **Start Application**
   ```bash
   go run ./cmd/api &
   go run ./cmd/worker &
   ```

5. **Monitor Health**
   ```bash
   curl http://localhost:8080/health
   ./scripts/monitor.sh check
   ```

## 📋 API Documentation

### Health Check Endpoints

#### GET `/health`
Comprehensive health check with detailed component status.

**Response:**
```json
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "database": {
      "status": "healthy|degraded|unhealthy",
      "message": "Status description",
      "last_checked": "2024-01-15T10:30:00Z",
      "response_time_ms": 12
    },
    "messaging": {
      "status": "healthy|degraded|unhealthy",
      "message": "Status description",
      "last_checked": "2024-01-15T10:30:00Z",
      "response_time_ms": 8
    }
  },
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

#### GET `/health/ready`
Kubernetes-style readiness probe.

**Response:**
```json
{
  "status": "ready|not ready",
  "reason": "Description if not ready"
}
```

#### GET `/health/live`
Kubernetes-style liveness probe.

**Response:**
```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Legacy Endpoints

#### GET `/api/health`
Legacy health check for backwards compatibility.

**Response:**
```json
{
  "status": "OK"
}
```

## 🚨 Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check database status
./scripts/db_manager.sh status

# Test connectivity
./scripts/test_runner.sh connectivity

# Check logs
tail -f logs/application.log
```

#### Messaging Issues
```bash
# Check RabbitMQ management
curl -u guest:guest http://localhost:15672/api/overview

# Check Service Bus connection
# Review connection string format

# Run messaging tests
./scripts/test_runner.sh connectivity
```

#### Health Check Issues
```bash
# Test health endpoints
curl -v http://localhost:8080/health
curl -v http://localhost:8080/health/ready
curl -v http://localhost:8080/health/live

# Check component health
./scripts/monitor.sh check
```

#### Container Issues
```bash
# Check container logs
docker-compose logs api
docker-compose logs worker

# Check container health
docker-compose ps
```

### Debugging Steps

1. **Check Configuration**
   ```bash
   ./scripts/test_runner.sh config
   ```

2. **Verify Connectivity**
   ```bash
   ./scripts/test_runner.sh connectivity
   ```

3. **Test Health Endpoints**
   ```bash
   ./scripts/test_runner.sh health
   ```

4. **Review Logs**
   ```bash
   tail -f logs/application.log
   tail -f logs/alerts.log
   ```

5. **Generate Report**
   ```bash
   ./scripts/monitor.sh report
   ./scripts/test_runner.sh report
   ```

## 🔐 Security

### Security Features

- **Non-root container execution** for enhanced security
- **SSL/TLS enforcement** for database connections
- **Managed identity support** for Azure resources
- **Credential encryption** and secure handling
- **Network security** with proper firewall rules
- **Health check security** with appropriate access controls

### Security Best Practices

- Use managed identities in Azure environments
- Enable SSL/TLS for all external connections
- Regularly rotate connection strings and passwords
- Monitor access logs and security events
- Use network policies to restrict access
- Implement proper RBAC for Azure resources

## 📊 Metrics & Observability

### Built-in Metrics

- **Response time metrics** for all health checks
- **Connection pool utilization** for database
- **System resource metrics** (CPU, memory, disk)
- **Application uptime** and availability
- **Error rates** and failure counts

### Integration Options

- **Prometheus** metrics export (can be added)
- **Grafana** dashboards for visualization
- **Azure Monitor** integration for cloud deployments
- **Log aggregation** with ELK stack or similar
- **Custom metrics** via webhook integrations

## 🎯 Next Steps

### Recommended Enhancements

1. **Metrics Export** - Add Prometheus metrics endpoint
2. **Distributed Tracing** - Implement OpenTelemetry
3. **Advanced Alerting** - Integrate with PagerDuty/OpsGenie
4. **Performance Profiling** - Add pprof endpoints
5. **Circuit Breaker** - Implement resilience patterns
6. **Rate Limiting** - Add API rate limiting
7. **Caching Layer** - Implement Redis caching
8. **Audit Logging** - Add comprehensive audit trails

### Production Checklist

- [ ] Configure monitoring and alerting
- [ ] Set up automated backups
- [ ] Implement log aggregation
- [ ] Configure SSL/TLS certificates
- [ ] Set up CI/CD pipelines
- [ ] Configure network security
- [ ] Implement RBAC policies
- [ ] Set up disaster recovery
- [ ] Configure auto-scaling policies
- [ ] Implement security scanning

## 📞 Support

For support and questions:

- **Health Check Issues**: Check `/health` endpoint and logs
- **Database Issues**: Use `./scripts/db_manager.sh status`
- **Messaging Issues**: Review connection configuration
- **Monitoring Issues**: Check `./scripts/monitor.sh check`
- **General Issues**: Run `./scripts/test_runner.sh all`

---

**🎉 Your application is now production-ready with enterprise-grade monitoring, health checks, and automation!**
