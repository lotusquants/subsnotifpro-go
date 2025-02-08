package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/messaging"
	"subsnotifpro-go/routes"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize database
	database.ConnectDatabase()   // Initialize the database
	database.AutoMigrateTables() // Auto-migrate tables

	// ✅ Get a **single** RabbitMQ Channel (initialized here)
	ch, err := messaging.GetChannel(ctx)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer messaging.CloseRabbitMQ()

	// ✅ Initialize RabbitMQ (Queues, Exchanges, Bindings)
	messaging.InitializeRabbitMQ(ch)

	// ✅ Start background services
	wg.Add(1)
	go func() {
		defer wg.Done()
		messaging.MonitorRabbitMQConnection(ctx)
	}()

	// ✅ Setup HTTP server
	router := routes.SetupRouter(ch)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	// ✅ Start HTTP Server (Non-Blocking)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🚀 Starting server on port %s...\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// ✅ Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // Wait for termination signal
	log.Println("🚦 Shutting down...")

	// ✅ Cancel background workers
	cancel()

	// ✅ Stop HTTP Server Gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ HTTP server shutdown error: %v", err)
	}

	// ✅ Wait for all goroutines to finish
	wg.Wait()

	// ✅ Close database
	database.CloseDatabase()

	// ✅ Cleanup RabbitMQ resources
	log.Println("🚦 Closing RabbitMQ Consumers...")
	messaging.CloseRabbitMQ()

	log.Println("✅ Server shutdown complete")
}
