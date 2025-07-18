#!/bin/bash

# Test Runner Script
# This script runs various tests for the application

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

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run all unit tests
    go test -v ./... -short
    
    if [ $? -eq 0 ]; then
        log_info "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Run connectivity tests
run_connectivity_tests() {
    log_info "Running connectivity tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run connectivity tests
    go test -v ./internal/tests -run TestMain
    
    if [ $? -eq 0 ]; then
        log_info "Connectivity tests passed"
        return 0
    else
        log_error "Connectivity tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests (excluding short tests)
    go test -v ./... -run Integration
    
    if [ $? -eq 0 ]; then
        log_info "Integration tests passed"
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Test application build
test_build() {
    log_info "Testing application build..."
    
    cd "$PROJECT_ROOT"
    
    # Test API build
    if go build -o ./build/api ./cmd/api; then
        log_info "API build successful"
    else
        log_error "API build failed"
        return 1
    fi
    
    # Test worker build
    if go build -o ./build/worker ./cmd/worker; then
        log_info "Worker build successful"
    else
        log_error "Worker build failed"
        return 1
    fi
    
    # Clean up build artifacts
    rm -rf ./build
    
    log_info "Build tests completed successfully"
    return 0
}

# Test configuration loading
test_config() {
    log_info "Testing configuration loading..."
    
    cd "$PROJECT_ROOT"
    
    # Test different configuration scenarios
    local test_configs=(
        "container:rabbitmq"
        "container:servicebus"
        "managed:servicebus"
        "external:rabbitmq"
    )
    
    for config in "${test_configs[@]}"; do
        IFS=':' read -ra ADDR <<< "$config"
        db_mode="${ADDR[0]}"
        msg_type="${ADDR[1]}"
        
        log_debug "Testing configuration: DB=$db_mode, MSG=$msg_type"
        
        # Set test environment variables
        export DB_DEPLOYMENT_MODE="$db_mode"
        export MESSAGING_TYPE="$msg_type"
        
        # Test configuration loading by building
        if go build -o /tmp/test_config ./cmd/api; then
            log_debug "Configuration test passed: $config"
        else
            log_error "Configuration test failed: $config"
            return 1
        fi
        
        # Clean up
        rm -f /tmp/test_config
    done
    
    log_info "Configuration tests completed successfully"
    return 0
}

# Run health check tests
test_health_checks() {
    log_info "Testing health check endpoints..."
    
    # Start the application in background
    cd "$PROJECT_ROOT"
    
    # Set test environment
    export DB_DEPLOYMENT_MODE="container"
    export MESSAGING_TYPE="rabbitmq"
    export SERVER_PORT="8081"
    
    # Start dependencies if needed
    if [ "$DB_DEPLOYMENT_MODE" = "container" ]; then
        log_debug "Starting database container for health check tests..."
        docker-compose --profile postgres up -d
    fi
    
    if [ "$MESSAGING_TYPE" = "rabbitmq" ]; then
        log_debug "Starting RabbitMQ container for health check tests..."
        docker-compose --profile rabbitmq up -d
    fi
    
    # Wait for containers to be ready
    sleep 10
    
    # Build and start API
    go build -o ./test_api ./cmd/api
    ./test_api &
    API_PID=$!
    
    # Wait for API to start
    sleep 5
    
    # Test health endpoints
    local endpoints=(
        "http://localhost:8081/health"
        "http://localhost:8081/health/ready"
        "http://localhost:8081/health/live"
        "http://localhost:8081/api/health"
    )
    
    local all_passed=true
    
    for endpoint in "${endpoints[@]}"; do
        log_debug "Testing endpoint: $endpoint"
        
        if curl -s -f "$endpoint" > /dev/null; then
            log_debug "✅ $endpoint responded successfully"
        else
            log_error "❌ $endpoint failed"
            all_passed=false
        fi
    done
    
    # Clean up
    kill $API_PID 2>/dev/null || true
    rm -f ./test_api
    
    # Stop containers
    docker-compose --profile postgres --profile rabbitmq down
    
    if [ "$all_passed" = true ]; then
        log_info "Health check tests passed"
        return 0
    else
        log_error "Health check tests failed"
        return 1
    fi
}

# Run all tests
run_all_tests() {
    log_info "Running all tests..."
    
    local all_passed=true
    
    # Run build tests
    if ! test_build; then
        all_passed=false
    fi
    
    # Run configuration tests
    if ! test_config; then
        all_passed=false
    fi
    
    # Run unit tests
    if ! run_unit_tests; then
        all_passed=false
    fi
    
    # Run connectivity tests
    if ! run_connectivity_tests; then
        all_passed=false
    fi
    
    # Run health check tests
    if ! test_health_checks; then
        all_passed=false
    fi
    
    if [ "$all_passed" = true ]; then
        log_info "🎉 All tests passed!"
        return 0
    else
        log_error "❌ Some tests failed"
        return 1
    fi
}

# Generate test report
generate_test_report() {
    local report_file="$PROJECT_ROOT/reports/test_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p "$(dirname "$report_file")"
    
    log_info "Generating test report: $report_file"
    
    {
        echo "=== Test Report ==="
        echo "Generated: $(date)"
        echo "Project: SubsNotifPro Go"
        echo ""
        
        echo "=== Test Results ==="
        if run_all_tests; then
            echo "Overall Status: PASSED"
        else
            echo "Overall Status: FAILED"
        fi
        echo ""
        
        echo "=== Environment ==="
        echo "Go Version: $(go version)"
        echo "OS: $(uname -s)"
        echo "Architecture: $(uname -m)"
        echo ""
        
        echo "=== Configuration ==="
        echo "DB Mode: ${DB_DEPLOYMENT_MODE:-not set}"
        echo "Messaging: ${MESSAGING_TYPE:-not set}"
        echo "Server Port: ${SERVER_PORT:-not set}"
        echo ""
        
    } > "$report_file"
    
    log_info "Test report generated: $report_file"
}

# Print usage information
usage() {
    echo "Test Runner Script"
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  unit                 Run unit tests"
    echo "  connectivity         Run connectivity tests"
    echo "  integration          Run integration tests"
    echo "  build                Test application build"
    echo "  config               Test configuration loading"
    echo "  health               Test health check endpoints"
    echo "  all                  Run all tests"
    echo "  report               Generate test report"
    echo
    echo "Environment Variables:"
    echo "  DB_DEPLOYMENT_MODE   Database deployment mode (container|managed|external)"
    echo "  MESSAGING_TYPE       Messaging type (rabbitmq|servicebus)"
    echo "  SERVER_PORT          Server port for testing (default: 8081)"
    echo
    echo "Examples:"
    echo "  $0 unit                      # Run unit tests"
    echo "  $0 connectivity              # Run connectivity tests"
    echo "  $0 all                       # Run all tests"
    echo "  $0 report                    # Generate test report"
}

# Main script logic
main() {
    local command="${1:-all}"
    
    load_env
    
    case $command in
        "unit")
            run_unit_tests
            ;;
        "connectivity")
            run_connectivity_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "build")
            test_build
            ;;
        "config")
            test_config
            ;;
        "health")
            test_health_checks
            ;;
        "all")
            run_all_tests
            ;;
        "report")
            generate_test_report
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
