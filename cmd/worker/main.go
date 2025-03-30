package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"subsnotifpro-go/database"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreSubscriptionRepository "subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"
	playstoreUserRepo "subsnotifpro-go/internal/playstore/user/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"
	userRepo "subsnotifpro-go/internal/users/repository"
	userService "subsnotifpro-go/internal/users/service"

	playstoreCatalogRepo "subsnotifpro-go/internal/playstore/products/repository"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"

	clientService "subsnotifpro-go/internal/playstore/client/service"

	"subsnotifpro-go/internal/playstore/rtdn/dispatch"
	rtdnRepo "subsnotifpro-go/internal/playstore/rtdn/repository"
	rtdnService "subsnotifpro-go/internal/playstore/rtdn/service"
	"subsnotifpro-go/queue"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ✅ Initialize RabbitMQ channel
	ch, err := queue.GetChannel(ctx)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer queue.CloseRabbitMQ() // Ensure RabbitMQ is closed when done

	// ✅ Initialize database and pass to repositories
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.CloseDatabase(db)

	// ✅ Pass only the interface (ServiceAccountProvider) to client
	clientService := clientService.NewPlaystoreClientService() // no cycle now

	// ✅ Now pass the client to apiService
	apiService := playstoreApiService.NewPlaystoreApiService(clientService)

	userRepo := userRepo.NewUserRepository()
	userService := userService.NewUserService(userRepo)

	psUserRepo := playstoreUserRepo.NewPlaystoreUserRepository()

	psUserService := playstoreUserService.NewPlaystoreUserService(userService, psUserRepo)

	psCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, 50)

	psCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psCatalogRepo, apiService)

	subscriptionRepo := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository()

	subscriptionService := playstoreSubscriptionService.NewPlaystoreSubscriptionService(db, subscriptionRepo, psUserService, apiService, psCatalogService)

	// ✅ Initialize the repositories and services
	rtdnRepo := rtdnRepo.NewRTDNRepository(db)                                                    // Adjust according to your repo
	rtdnService := rtdnService.NewRTDNService(ctx, rtdnRepo, apiService, subscriptionService, db) // Adjust according to your service

	// ✅ Create and start the consumer
	consumer := dispatch.NewConsumer(ch, rtdnRepo, rtdnService)
	go consumer.Start(ctx)

	go dispatch.StartDLQConsumer(ctx, ch)

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
