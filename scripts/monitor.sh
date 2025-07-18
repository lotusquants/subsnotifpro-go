#!/bin/bash

# Application Monitoring Script
# This script provides monitoring and alerting for the application in different deployment modes

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Load environment variables
load_env() {
    if [ -f "$PROJECT_ROOT/.env" ]; then
        # Process .env file and export variables, handling inline comments
        while IFS='=' read -r key value; do
            # Skip empty lines and comments
            if [[ -z "$key" || "$key" =~ ^#.* ]]; then
                continue
            fi
            # Remove inline comments from value
            value=$(echo "$value" | sed 's/#.*//' | xargs)
            # Export the variable
            export "$key"="$value"
        done < <(grep -v '^#' "$PROJECT_ROOT/.env" | grep -v '^$')
        log_info "Environment variables loaded from .env file"
    else
        log_warn "No .env file found, using system environment variables"
    fi
}

# Configuration variables
API_URL=${API_URL:-"http://localhost:8080"}
HEALTH_ENDPOINT=${HEALTH_ENDPOINT:-"/health"}
READINESS_ENDPOINT=${READINESS_ENDPOINT:-"/health/ready"}
LIVENESS_ENDPOINT=${LIVENESS_ENDPOINT:-"/health/live"}
WEBHOOK_URL=${WEBHOOK_URL:-""}
ALERT_EMAIL=${ALERT_EMAIL:-""}
CHECK_INTERVAL=${CHECK_INTERVAL:-60}
TIMEOUT=${TIMEOUT:-10}

# Health check status
LAST_STATUS=""
FAILURE_COUNT=0
MAX_FAILURES=${MAX_FAILURES:-3}

# Check API health
check_api_health() {
    local endpoint="$1"
    local expected_status="${2:-200}"
    
    log_debug "Checking API health at: $API_URL$endpoint"
    
    local response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json --connect-timeout "$TIMEOUT" "$API_URL$endpoint" 2>/dev/null || echo "000")
    
    if [ "$response" = "$expected_status" ]; then
        log_info "API health check passed: $endpoint"
        return 0
    else
        log_error "API health check failed: $endpoint (HTTP $response)"
        return 1
    fi
}

# Check database connectivity
check_database() {
    local mode=${DB_DEPLOYMENT_MODE:-"container"}
    
    case $mode in
        "container")
            DB_HOST=${DB_HOST:-"localhost"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"postgres"}
            DB_PASSWORD=${DB_PASSWORD:-"postgres"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            ;;
        "managed")
            DB_HOST=${DB_HOST:-"your-server.postgres.database.azure.com"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"your-user"}
            DB_PASSWORD=${DB_PASSWORD:-"your-password"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            ;;
        "external")
            DB_HOST=${DB_HOST:-"localhost"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"postgres"}
            DB_PASSWORD=${DB_PASSWORD:-"postgres"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            ;;
    esac
    
    log_debug "Checking database connectivity: $DB_HOST:$DB_PORT/$DB_NAME"
    
    export PGPASSWORD="$DB_PASSWORD"
    if timeout "$TIMEOUT" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        log_info "Database connectivity check passed"
        return 0
    else
        log_error "Database connectivity check failed"
        return 1
    fi
}

# Check messaging backend
check_messaging() {
    local messaging_type=${MESSAGING_TYPE:-"rabbitmq"}
    
    case $messaging_type in
        "rabbitmq")
            check_rabbitmq
            ;;
        "servicebus")
            check_servicebus
            ;;
        *)
            log_error "Unknown messaging type: $messaging_type"
            return 1
            ;;
    esac
}

# Check RabbitMQ connectivity
check_rabbitmq() {
    local host=${RABBITMQ_HOST:-"localhost"}
    local port=${RABBITMQ_PORT:-"5672"}
    local user=${RABBITMQ_USERNAME:-"guest"}
    local pass=${RABBITMQ_PASSWORD:-"guest"}
    
    log_debug "Checking RabbitMQ connectivity: $host:$port"
    
    # Check if RabbitMQ management API is accessible
    local management_port=${RABBITMQ_MANAGEMENT_PORT:-"15672"}
    local management_url="http://$host:$management_port/api/overview"
    
    if curl -s -u "$user:$pass" --connect-timeout "$TIMEOUT" "$management_url" > /dev/null 2>&1; then
        log_info "RabbitMQ connectivity check passed"
        return 0
    else
        log_error "RabbitMQ connectivity check failed"
        return 1
    fi
}

# Check Azure Service Bus connectivity
check_servicebus() {
    local connection_string=${SERVICEBUS_CONNECTION_STRING:-""}
    
    if [ -z "$connection_string" ]; then
        log_error "Service Bus connection string not configured"
        return 1
    fi
    
    log_debug "Checking Service Bus connectivity"
    
    # This is a simplified check - in practice, you might want to use Azure CLI or SDK
    # For now, we'll just check if the connection string is properly formatted
    if echo "$connection_string" | grep -q "Endpoint=sb://"; then
        log_info "Service Bus connectivity check passed"
        return 0
    else
        log_error "Service Bus connectivity check failed"
        return 1
    fi
}

# Check Docker containers (if running in container mode)
check_containers() {
    local mode=${DB_DEPLOYMENT_MODE:-"container"}
    local messaging_type=${MESSAGING_TYPE:-"rabbitmq"}
    
    if [ "$mode" = "container" ] || [ "$messaging_type" = "rabbitmq" ]; then
        log_debug "Checking Docker containers"
        
        # Check if docker is running
        if ! docker info > /dev/null 2>&1; then
            log_error "Docker is not running"
            return 1
        fi
        
        # Check postgres container if using container mode
        if [ "$mode" = "container" ]; then
            if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "postgres.*Up"; then
                log_info "PostgreSQL container is running"
            else
                log_error "PostgreSQL container is not running"
                return 1
            fi
        fi
        
        # Check RabbitMQ container if using RabbitMQ
        if [ "$messaging_type" = "rabbitmq" ]; then
            if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "rabbitmq.*Up"; then
                log_info "RabbitMQ container is running"
            else
                log_error "RabbitMQ container is not running"
                return 1
            fi
        fi
    fi
    
    return 0
}

# Check application processes
check_processes() {
    log_debug "Checking application processes"
    
    # Check API process
    if pgrep -f "subsnotifpro.*api" > /dev/null; then
        log_info "API process is running"
    else
        log_warn "API process is not running"
    fi
    
    # Check worker process
    if pgrep -f "subsnotifpro.*worker" > /dev/null; then
        log_info "Worker process is running"
    else
        log_warn "Worker process is not running"
    fi
}

# Get system metrics
get_system_metrics() {
    log_debug "Collecting system metrics"
    
    # CPU usage
    local cpu_usage=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | cut -d'%' -f1)
    echo "CPU Usage: ${cpu_usage}%"
    
    # Memory usage
    local memory_info=$(vm_stat | grep "Pages free\|Pages active\|Pages inactive\|Pages speculative\|Pages wired down")
    echo "Memory Info:"
    echo "$memory_info"
    
    # Disk usage
    local disk_usage=$(df -h / | tail -1 | awk '{print $5}')
    echo "Disk Usage: $disk_usage"
    
    # Load average
    local load_avg=$(uptime | awk '{print $10, $11, $12}')
    echo "Load Average: $load_avg"
}

# Send alert
send_alert() {
    local subject="$1"
    local message="$2"
    
    log_warn "Sending alert: $subject"
    
    # Send webhook notification
    if [ -n "$WEBHOOK_URL" ]; then
        curl -X POST -H "Content-Type: application/json" \
             -d "{\"text\":\"$subject\n$message\"}" \
             "$WEBHOOK_URL" > /dev/null 2>&1
    fi
    
    # Send email notification
    if [ -n "$ALERT_EMAIL" ]; then
        echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
    fi
    
    # Log to file
    local log_file="$PROJECT_ROOT/logs/alerts.log"
    mkdir -p "$(dirname "$log_file")"
    echo "$(date): $subject - $message" >> "$log_file"
}

# Comprehensive health check
run_health_check() {
    log_info "Running comprehensive health check..."
    
    local all_checks_passed=true
    
    # Check API health
    if ! check_api_health "$HEALTH_ENDPOINT"; then
        all_checks_passed=false
    fi
    
    # Check database
    if ! check_database; then
        all_checks_passed=false
    fi
    
    # Check messaging
    if ! check_messaging; then
        all_checks_passed=false
    fi
    
    # Check containers (if applicable)
    if ! check_containers; then
        all_checks_passed=false
    fi
    
    # Check processes
    check_processes
    
    # Get system metrics
    get_system_metrics
    
    # Handle status change
    local current_status
    if [ "$all_checks_passed" = true ]; then
        current_status="healthy"
        FAILURE_COUNT=0
    else
        current_status="unhealthy"
        ((FAILURE_COUNT++))
    fi
    
    # Send alert if status changed or failure threshold reached
    if [ "$LAST_STATUS" != "$current_status" ] || [ $FAILURE_COUNT -ge $MAX_FAILURES ]; then
        local subject="Application Health Status: $current_status"
        local message="Application health check status: $current_status\nFailure count: $FAILURE_COUNT\nTimestamp: $(date)"
        send_alert "$subject" "$message"
        LAST_STATUS="$current_status"
    fi
    
    return $([ "$all_checks_passed" = true ] && echo 0 || echo 1)
}

# Monitor continuously
monitor() {
    log_info "Starting continuous monitoring (interval: ${CHECK_INTERVAL}s)"
    
    while true; do
        run_health_check
        sleep "$CHECK_INTERVAL"
    done
}

# Generate monitoring report
generate_report() {
    local report_file="$PROJECT_ROOT/reports/monitoring_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p "$(dirname "$report_file")"
    
    log_info "Generating monitoring report: $report_file"
    
    {
        echo "=== Application Monitoring Report ==="
        echo "Generated: $(date)"
        echo "Configuration:"
        echo "  API URL: $API_URL"
        echo "  DB Mode: ${DB_DEPLOYMENT_MODE:-container}"
        echo "  Messaging: ${MESSAGING_TYPE:-rabbitmq}"
        echo ""
        
        echo "=== Health Check Results ==="
        run_health_check
        echo ""
        
        echo "=== System Metrics ==="
        get_system_metrics
        echo ""
        
        echo "=== Docker Status ==="
        if command -v docker &> /dev/null; then
            docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        else
            echo "Docker not available"
        fi
        echo ""
        
        echo "=== Application Logs (last 20 lines) ==="
        if [ -f "$PROJECT_ROOT/logs/application.log" ]; then
            tail -20 "$PROJECT_ROOT/logs/application.log"
        else
            echo "No application logs found"
        fi
        
    } > "$report_file"
    
    log_info "Report generated: $report_file"
}

# Setup monitoring service
setup_monitoring() {
    log_info "Setting up monitoring service..."
    
    # Create systemd service file for monitoring
    local service_file="/etc/systemd/system/subsnotifpro-monitor.service"
    
    cat > "$service_file" << EOF
[Unit]
Description=SubsNotifPro Monitoring Service
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=$PROJECT_ROOT
ExecStart=$SCRIPT_DIR/monitor.sh monitor
Restart=always
RestartSec=30
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start the service
    systemctl daemon-reload
    systemctl enable subsnotifpro-monitor
    systemctl start subsnotifpro-monitor
    
    log_info "Monitoring service setup completed"
}

# Print usage information
usage() {
    echo "Application Monitoring Script"
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  check               Run single health check"
    echo "  monitor             Start continuous monitoring"
    echo "  report              Generate monitoring report"
    echo "  setup               Setup monitoring as a service"
    echo
    echo "Environment Variables:"
    echo "  API_URL             API base URL (default: http://localhost:8080)"
    echo "  HEALTH_ENDPOINT     Health check endpoint (default: /health)"
    echo "  WEBHOOK_URL         Webhook URL for alerts"
    echo "  ALERT_EMAIL         Email address for alerts"
    echo "  CHECK_INTERVAL      Check interval in seconds (default: 60)"
    echo "  TIMEOUT             Request timeout in seconds (default: 10)"
    echo "  MAX_FAILURES        Max failures before alert (default: 3)"
    echo
    echo "Examples:"
    echo "  $0 check                     # Run single health check"
    echo "  $0 monitor                   # Start continuous monitoring"
    echo "  $0 report                    # Generate monitoring report"
    echo "  CHECK_INTERVAL=30 $0 monitor # Monitor every 30 seconds"
}

# Main script logic
main() {
    local command="${1:-check}"
    
    case $command in
        "check")
            load_env
            run_health_check
            ;;
        "monitor")
            load_env
            monitor
            ;;
        "report")
            load_env
            generate_report
            ;;
        "setup")
            setup_monitoring
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
