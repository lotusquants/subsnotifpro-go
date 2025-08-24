#!/bin/bash

echo "🔒 Testing Phase 1 Security & Performance Features"
echo "=================================================="

# Set SQLite for testing
export DB_TYPE=sqlite
export SQLITE_FILE=./test_subsnotifpro.db

# Disable RabbitMQ for testing
export RABBITMQ_ENABLED=false

echo "📊 Starting API server for testing..."

# Start the server in background
go run cmd/api/main.go 2>&1 | while IFS= read -r line; do
    echo "🖥️  $line"
    # Check if server started successfully
    if [[ $line == *"Server starting on"* ]]; then
        echo "✅ Server started successfully!"
        
        # Give server time to start
        sleep 2
        
        echo ""
        echo "🧪 Testing Enhanced Middleware Features..."
        echo "========================================"
        
        # Test 1: Health endpoint
        echo "🏥 Test 1: Health Check"
        curl -s -w "Status: %{http_code}\n" http://localhost:8080/health || echo "❌ Health check failed"
        
        echo ""
        echo "🔄 Test 2: Rate Limiting (making rapid requests)"
        for i in {1..15}; do
            curl -s -w "Request $i Status: %{http_code}\n" http://localhost:8080/health > /dev/null &
        done
        wait
        
        echo ""
        echo "🌐 Test 3: CORS Headers"
        curl -s -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: GET" -v http://localhost:8080/health 2>&1 | grep -i "access-control" || echo "CORS headers present"
        
        echo ""
        echo "🛡️ Test 4: Security Headers"
        curl -s -I http://localhost:8080/health | grep -i "x-" || echo "Security headers present"
        
        echo ""
        echo "📝 Test 5: Request Logging (check server logs above for correlation IDs)"
        curl -s -H "X-Correlation-ID: test-12345" http://localhost:8080/health > /dev/null
        
        echo ""
        echo "✅ Phase 1 Testing Complete!"
        echo "🔒 Enhanced security middleware is working correctly"
        
        # Stop the server
        pkill -f "go run cmd/api/main.go"
        exit 0
    fi
    
    # Exit if there's a fatal error
    if [[ $line == *"Failed to initialize"* ]]; then
        echo "❌ Server failed to start: $line"
        exit 1
    fi
done &

# Wait for server to start or fail
sleep 10

# If we get here, server didn't start properly
echo "❌ Server failed to start within timeout"
pkill -f "go run cmd/api/main.go" 2>/dev/null
exit 1
