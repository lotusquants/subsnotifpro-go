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
	"subsnotifpro-go/internal/google_playstore/rtdn"
	"subsnotifpro-go/internal/google_playstore/rtdn/queue"
	"subsnotifpro-go/internal/messaging"
	"subsnotifpro-go/internal/metrics"
	"subsnotifpro-go/routes"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize database
	database.ConnectDatabase()
	database.AutoMigrateTables()

	// ✅ Get a **single** RabbitMQ Channel
	ch, err := messaging.GetChannel(ctx)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer func() {
		log.Println("🚦 Closing RabbitMQ connection...")
		_ = ch.Close()
	}()

	// ✅ Initialize RabbitMQ (Queues, Exchanges, Bindings)
	messaging.InitializeRabbitMQ(ch)

	// ✅ Start background services
	wg.Add(1)
	go func() {
		defer wg.Done()
		messaging.MonitorRabbitMQConnection(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		metrics.StartMetricsServer(ctx)
	}()

	// ✅ Start Queue Consumers with a shared channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.StartQueueConsumer(ctx, ch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.StartDLQConsumer(ctx, ch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rtdn.ProcessPendingEvents(ctx, 10)
	}()

	// ✅ Setup HTTP server
	router := routes.SetupRouter()
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

	// // ✅ Close database
	database.CloseDatabase()

	// ✅ Cleanup RabbitMQ resources
	log.Println("🚦 Closing RabbitMQ Consumers...")
	messaging.CloseRabbitMQ()

	log.Println("✅ Server shutdown complete")
}
