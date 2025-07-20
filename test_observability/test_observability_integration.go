// Package main provides comprehensive testing for the observability implementation
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/container"
	"subsnotifpro-go/routes"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 Testing Complete Observability Implementation")

	// Load configuration
	cfg := config.LoadConfig()

	// Create container with observability
	containerInstance, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create container: %v", err)
	}
	defer containerInstance.Close()

	// Validate observability components
	fmt.Println("\n📊 Validating Observability Components:")
	
	if containerInstance.MetricsRegistry != nil {
		fmt.Println("✅ Metrics Registry: Initialized")
	} else {
		fmt.Println("❌ Metrics Registry: Not initialized")
	}
	
	if containerInstance.TracerProvider != nil {
		fmt.Println("✅ Tracer Provider: Initialized")
		
		// Test tracing capabilities
		ctx := context.Background()
		ctx, span := containerInstance.TracerProvider.StartSpan(ctx, "test.operation")
		containerInstance.TracerProvider.SetAttributes(ctx, map[string]interface{}{
			"test.attribute": "test_value",
			"test.number":    42,
		})
		span.End()
		
		fmt.Printf("   - Sample Trace ID: %s\n", containerInstance.TracerProvider.GetTraceID(ctx))
	} else {
		fmt.Println("⚠️  Tracer Provider: Not initialized (tracing may be disabled)")
	}
	
	if containerInstance.ObservabilityMiddleware != nil {
		fmt.Println("✅ Observability Middleware: Initialized")
	} else {
		fmt.Println("❌ Observability Middleware: Not initialized")
	}

	// Create router with observability
	deps := containerInstance.GetRouteDependencies()
	router := routes.SetupRouter(deps)

	// Add test routes to demonstrate observability
	addObservabilityTestRoutes(router, deps)

	fmt.Println("\n🌐 Starting Test Server with Observability:")
	fmt.Println("   - API Server: http://localhost:8080")
	fmt.Println("   - Metrics: http://localhost:9090/metrics")
	fmt.Println("   - Health: http://localhost:8080/health")
	
	fmt.Println("\n🧪 Test Endpoints:")
	fmt.Println("   - GET /observability/test/subscription?platform=playstore")
	fmt.Println("   - GET /observability/test/database")
	fmt.Println("   - GET /observability/test/message/rtdn_queue/subscription_event")
	fmt.Println("   - GET /observability/test/error (for error testing)")

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("\n⏳ Starting server... (Press Ctrl+C to stop)")
	
	// Simulate some test requests
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n🔄 Running automated tests...")
		runAutomatedTests()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed: %v", err)
	}
}

// addObservabilityTestRoutes adds test routes to demonstrate observability features
func addObservabilityTestRoutes(router *gin.Engine, deps *routes.RouteDependencies) {
	if deps.ObservabilityMiddleware == nil {
		fmt.Println("⚠️  No observability middleware available for test routes")
		return
	}

	testGroup := router.Group("/observability/test")
	
	// Test business operation tracing
	testGroup.GET("/subscription", func(c *gin.Context) {
		platform := c.DefaultQuery("platform", "unknown")
		
		// Start business operation with observability
		_, finish := deps.ObservabilityMiddleware.BusinessOperationMiddleware("subscription.test", map[string]interface{}{
			"platform": platform,
			"test":     true,
		})(c.Request.Context())
		
		// Simulate business logic
		time.Sleep(100 * time.Millisecond)
		
		// Finish observability
		finish(nil)
		
		c.JSON(200, gin.H{
			"status":   "success",
			"platform": platform,
			"message":  "Test subscription processing with observability",
			"trace_id": deps.ObservabilityMiddleware != nil,
		})
	})
	
	// Test database operation tracing
	testGroup.GET("/database", func(c *gin.Context) {
		ctx, finish := deps.ObservabilityMiddleware.DatabaseMiddleware("SELECT", "test_table")(c.Request.Context())
		
		// Simulate database query
		time.Sleep(50 * time.Millisecond)
		
		finish()
		
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Test database operation with observability",
			"context": ctx != nil,
		})
	})
	
	// Test message processing tracing
	testGroup.GET("/message/:queue/:type", func(c *gin.Context) {
		queue := c.Param("queue")
		messageType := c.Param("type")
		
		_, finish := deps.ObservabilityMiddleware.MessageProcessingMiddleware(queue, messageType)(c.Request.Context())
		
		// Simulate message processing
		time.Sleep(75 * time.Millisecond)
		
		finish(nil)
		
		c.JSON(200, gin.H{
			"status":       "success",
			"queue":        queue,
			"message_type": messageType,
			"message":      "Test message processing with observability",
			"context":      true,
		})
	})
	
	// Test error handling in observability
	testGroup.GET("/error", func(c *gin.Context) {
		_, finish := deps.ObservabilityMiddleware.BusinessOperationMiddleware("test.error", map[string]interface{}{
			"will_fail": true,
		})(c.Request.Context())
		
		// Simulate error
		time.Sleep(25 * time.Millisecond)
		err := fmt.Errorf("simulated error for testing")
		
		finish(err)
		
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "Test error handling with observability",
			"error":   err.Error(),
			"context": true,
		})
	})
}

// runAutomatedTests performs automated testing of observability endpoints
func runAutomatedTests() {
	baseURL := "http://localhost:8080"
	
	testEndpoints := []string{
		"/observability/test/subscription?platform=playstore",
		"/observability/test/database",
		"/observability/test/message/rtdn_queue/subscription_event", 
		"/observability/test/error",
		"/health",
		"/health/ready",
		"/health/live",
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, endpoint := range testEndpoints {
		url := baseURL + endpoint
		fmt.Printf("   Testing: %s\n", url)
		
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("      ❌ Error: %v\n", err)
			continue
		}
		
		fmt.Printf("      ✅ Status: %d\n", resp.StatusCode)
		resp.Body.Close()
		
		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println("\n📈 Check metrics at: http://localhost:9090/metrics")
	fmt.Println("🔍 Observability test completed!")
}
