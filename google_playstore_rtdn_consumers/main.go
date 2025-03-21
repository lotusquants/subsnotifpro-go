package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"subsnotifpro-go/internal/database"
	clientService "subsnotifpro-go/internal/google_playstore/client/service"
	"subsnotifpro-go/internal/google_playstore/rtdn/queue"
	rtdnRepo "subsnotifpro-go/internal/google_playstore/rtdn/repository"
	rtdnService "subsnotifpro-go/internal/google_playstore/rtdn/service"
	settingsRepo "subsnotifpro-go/internal/google_playstore/settings/repository"
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

	// ✅ Initialize database and pass to repositories
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.CloseDatabase(db)

	settingsRepo := settingsRepo.NewPlaystoreSettingsRepository()
	clientService := clientService.NewGooglePlayClientService(ctx, settingsRepo, db)

	// ✅ Initialize the repositories and services
	rtdnRepo := rtdnRepo.NewRTDNRepository(db)                              // Adjust according to your repo
	rtdnService := rtdnService.NewRTDNService(ctx, rtdnRepo, clientService) // Adjust according to your service

	// ✅ Create and start the consumer
	consumer := queue.NewConsumer(ch, rtdnRepo, rtdnService)
	go consumer.Start(ctx)

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
