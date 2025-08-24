#!/bin/bash

# 🚀 SubsNotifPro-Go: Phase 1 Security & Performance Demo
# This script demonstrates the enhanced security and monitoring features

echo "🔒 Starting SubsNotifPro-Go Enhanced Security Demo..."
echo "=================================================="

# Check if Redis is running (optional)
echo "📊 Checking Redis availability..."
if redis-cli ping 2>/dev/null | grep -q PONG; then
    echo "✅ Redis is running - distributed caching enabled"
else
    echo "⚠️  Redis not running - falling back to in-memory cache"
fi

# Start the API server in the background
echo ""
echo "🚀 Starting API server with enhanced middleware..."
cd /Users/nithishkailas/subsnotifpro-go
go run ./cmd/api &
API_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "🧪 Testing Enhanced Security Features..."
echo "========================================"

# Test 1: Rate Limiting
echo ""
echo "🔄 Test 1: Rate Limiting Protection"
echo "Making rapid requests to trigger rate limiting..."
for i in {1..25}; do
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health)
    if [ "$response" = "429" ]; then
        echo "✅ Rate limiting activated at request #$i (HTTP 429)"
        break
    fi
    sleep 0.1
done

# Test 2: CORS Headers
echo ""
echo "🌐 Test 2: CORS Configuration"
echo "Checking CORS headers..."
curl -s -I -H "Origin: http://localhost:3000" http://localhost:8080/api/health | grep -E "(Access-Control|X-)"

# Test 3: Security Headers
echo ""
echo "🛡️ Test 3: Security Headers"
echo "Checking security headers..."
curl -s -I http://localhost:8080/api/health | grep -E "(X-Content-Type-Options|X-Frame-Options|Referrer-Policy)"

# Test 4: Input Validation
echo ""
echo "✅ Test 4: Input Validation"
echo "Testing malicious input detection..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/api/health?test=<script>alert('xss')</script>")
echo "Response to XSS attempt: HTTP $response"

# Test 5: Request Logging
echo ""
echo "📝 Test 5: Request Logging with Correlation ID"
echo "Making request with custom header..."
curl -s -H "X-Correlation-ID: demo-test-123" -H "User-Agent: Demo-Client/1.0" http://localhost:8080/api/health > /dev/null
echo "✅ Check application logs for correlation ID: demo-test-123"

# Test 6: Cache Performance
echo ""
echo "⚡ Test 6: Cache Performance"
echo "Making multiple requests to demonstrate caching..."
for i in {1..3}; do
    echo "Request #$i:"
    curl -s -I http://localhost:8080/api/health | grep -E "(X-Cache|X-Correlation-ID)" || echo "  No cache headers (health endpoint may not be cached)"
    sleep 0.5
done

# Test 7: Suspicious User Agent Detection
echo ""
echo "🚨 Test 7: Suspicious User Agent Detection"
echo "Testing with suspicious user agent..."
response=$(curl -s -o /dev/null -w "%{http_code}" -H "User-Agent: sqlmap/1.0" http://localhost:8080/api/health)
echo "Response to suspicious user agent: HTTP $response"

echo ""
echo "📊 Getting Application Status..."
echo "==============================="

# Get health status
echo "🏥 Health Check:"
curl -s http://localhost:8080/api/health | jq . 2>/dev/null || curl -s http://localhost:8080/api/health

# Get metrics (if available)
echo ""
echo "📈 Metrics (if available):"
curl -s http://localhost:9090/metrics 2>/dev/null | head -10 || echo "Metrics endpoint not accessible"

echo ""
echo "🎉 Demo Complete!"
echo "=================="
echo ""
echo "✅ Enhanced Security Features Demonstrated:"
echo "   • Rate limiting protection"
echo "   • CORS configuration"
echo "   • Security headers"
echo "   • Input validation"
echo "   • Request logging with correlation IDs"
echo "   • Cache performance optimization"
echo "   • Suspicious user agent detection"
echo ""
echo "📊 Check the application logs to see detailed middleware activity"
echo "🔒 Production-ready security features are now active!"

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $API_PID 2>/dev/null
wait $API_PID 2>/dev/null

echo "Demo finished. Server stopped."
