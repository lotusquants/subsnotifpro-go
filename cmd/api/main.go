package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subsnotifpro-go/internal/constants"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	dispatch "subsnotifpro-go/internal/playstore/dispatch"

	playstoreApiHandler "subsnotifpro-go/internal/playstore/api/handler"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreClientService "subsnotifpro-go/internal/playstore/client/service"

	playstoreCatalogHandler "subsnotifpro-go/internal/playstore/products/handler"
	playstoreCatalogRepo "subsnotifpro-go/internal/playstore/products/repository"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"
	playstoreRtdnHandler "subsnotifpro-go/internal/playstore/rtdn/handler"
	playstoreRTDNRepo "subsnotifpro-go/internal/playstore/rtdn/repository"
	playstoreRTDNService "subsnotifpro-go/internal/playstore/rtdn/service"
	playstoreSettingsHandler "subsnotifpro-go/internal/playstore/settings/handler"
	playstoreSettingRepo "subsnotifpro-go/internal/playstore/settings/repository"
	playstoreSettingServicePkg "subsnotifpro-go/internal/playstore/settings/service"
	playstoreSubscriptionRepository "subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"
	playstoreUserRepo "subsnotifpro-go/internal/playstore/user/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"
	userRepo "subsnotifpro-go/internal/users/repository"
	userService "subsnotifpro-go/internal/users/service"
	"sync"
	"syscall"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/migrations"
	queue "subsnotifpro-go/queue"

	"subsnotifpro-go/routes"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// // Initialize logger with options
	// logger := logger.New(
	// 	logger.WithLevel(logger.DebugLevel),
	// 	logger.WithCaller(true),
	// )

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize database and pass to repositories
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	database.AutoMigrateTables(db)       // Auto-migrate tables
	migrations.ApplyCompositeIndexes(db) // Apply the composite indexes

	// ✅ Get a **single** RabbitMQ Channel (initialized here)
	ch, err := queue.GetChannel(ctx)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer queue.CloseRabbitMQ()

	// ✅ Initialize RabbitMQ (Queues, Exchanges, Bindings)
	queue.InitializeRabbitMQ(ch)

	// ✅ Start background services
	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.MonitorRabbitMQConnection(ctx)
	}()

	// ✅ Initialize Repositories, Services and Handlers

	psRtdnRepo := playstoreRTDNRepo.NewRTDNRepository(db)
	psSettingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository(db)
	psSubscriptionCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, getSyncBatchSize())

	// 🟢 Services

	// Create generic publisher
	msgPublisher := messaging.NewRabbitMQPublisher(messaging.PublisherOptions{
		Channel: ch,
		// Metrics: metrics.NewPublisherMetrics(),
	})

	// Create domain-specific publishers
	googlePlayPublisher := dispatch.NewGooglePlayPublisher(msgPublisher)

	// ✅ Create a temporary placeholder for settingsService (declare first)
	// ✅ Step 1: Declare placeholder
	psClientService := playstoreClientService.NewPlaystoreClientService()
	psApiService := playstoreApiService.NewPlaystoreApiService(psClientService)

	userRepo := userRepo.NewUserRepository()
	userService := userService.NewUserService(userRepo)

	psUserRepo := playstoreUserRepo.NewPlaystoreUserRepository()

	psUserService := playstoreUserService.NewPlaystoreUserService(userService, psUserRepo)

	psCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, 50)

	psCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psCatalogRepo, psApiService)

	subscriptionRepo := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository()

	subscriptionService := playstoreSubscriptionService.NewPlaystoreSubscriptionService(db, subscriptionRepo, psUserService, psApiService, psCatalogService)

	psRtdnService := playstoreRTDNService.NewRTDNService(ctx, psRtdnRepo, psApiService, subscriptionService, db, googlePlayPublisher)
	psSettingsService := playstoreSettingServicePkg.NewPlaystoreSettingsService(psSettingsRepo, psApiService, db)
	psClientService.SetAccountProvider(psSettingsService) // this works via interface

	// ✅ Step 4: Use the same instance
	psRtdnHandler := playstoreRtdnHandler.NewRTDNHandler(psRtdnService)
	psSettingsHandler := playstoreSettingsHandler.NewPlaystoreSettingsHandler(psSettingsService)
	psApiHandler := playstoreApiHandler.NewPlaystoreClientHandler(psApiService)
	psSubscriptionCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psSubscriptionCatalogRepo, psApiService)
	psSubscriptionCatalogHandler := playstoreCatalogHandler.NewSubscriptionCatalogHandler(psSubscriptionCatalogService)

	deps := &routes.RouteDependencies{
		PlaystoreRTDNHandler:                psRtdnHandler,
		PlaystoreSettingsHandler:            psSettingsHandler,
		PlaystoreApiHandler:                 psApiHandler,
		PlaystoreSubscriptionCatalogHandler: psSubscriptionCatalogHandler,
	}

	router := routes.SetupRouter(deps)

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
	database.CloseDatabase(db)

	// ✅ Cleanup RabbitMQ resources
	log.Println("🚦 Closing RabbitMQ Consumers...")
	queue.CloseRabbitMQ()

	log.Println("✅ Server shutdown complete")
}

func getSyncBatchSize() int {
	if envSize := os.Getenv("SYNC_BATCH_SIZE"); envSize != "" {
		if parsedSize, err := strconv.Atoi(envSize); err == nil {
			return parsedSize
		}
	}
	return constants.DEFAULT_SYNC_BATCH_SIZE
}
