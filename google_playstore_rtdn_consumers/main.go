package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/google_playstore/rtdn/queue"
	"subsnotifpro-go/internal/messaging"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ✅ Initialize RabbitMQ channel
	ch, err := messaging.GetChannel(ctx)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer messaging.CloseRabbitMQ() // Ensure RabbitMQ is closed when done

	// ✅ Initialize the database for the consumer (No return value)
	database.ConnectDatabase()     // Establish DB connection (doesn't return anything)
	defer database.CloseDatabase() // Ensure DB connection is closed when done

	// ✅ Start the consumer with the channel
	go queue.StartQueueConsumer(ctx, ch)
	go queue.StartDLQConsumer(ctx, ch)

	// ✅ Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // Wait for termination signal
	log.Println("🚦 Shutting down...")

	// ✅ Cancel background workers
	cancel()

	// ✅ Wait for all goroutines to finish (you could use a WaitGroup here if needed)
	log.Println("✅ Consumer service shutdown complete")
}
