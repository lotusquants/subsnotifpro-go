package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"subsnotifpro-go/cmd/api/app"
	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/container"
)

// BuildInfo contains build-time information
var BuildInfo = app.BuildInfo{
	Version:   getEnvOrDefault("VERSION", "1.0.0"),
	GitCommit: getEnvOrDefault("GIT_COMMIT", "unknown"),
	BuildTime: getEnvOrDefault("BUILD_TIME", "unknown"),
	GoVersion: getEnvOrDefault("GO_VERSION", "unknown"),
}

func main() {
	// Create root context
	ctx := context.Background()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize container with all dependencies
	container, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("❌ Error closing container: %v", err)
		}
	}()

	// Get route dependencies from container
	deps := container.GetRouteDependencies()

	// Create application with framework
	application, err := app.NewApplication(
		ctx,
		BuildInfo,
		app.WithDependencies(deps),
		app.WithHealthCheck(func(ctx context.Context) error {
			if deps.HealthChecker != nil {
				health := deps.HealthChecker.CheckDatabase(ctx)
				if health.Status != "healthy" {
					return fmt.Errorf("database health check failed: %s", health.Message)
				}
			}
			return nil
		}),
		app.WithShutdownFunc(func() error {
			return container.Close()
		}),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create application: %v", err)
	}

	// Initialize and run the application
	if err := application.Initialize(ctx); err != nil {
		log.Fatalf("❌ Failed to initialize application: %v", err)
	}

	// Run the application (blocks until shutdown)
	if err := application.Run(ctx); err != nil {
		log.Fatalf("❌ Application error: %v", err)
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
