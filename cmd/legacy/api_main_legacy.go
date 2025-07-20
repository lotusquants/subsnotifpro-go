package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	appStoreDistpatch "subsnotifpro-go/internal/appstore/dispatch"
	appStoreSettingsHandler "subsnotifpro-go/internal/appstore/settings/handler"
	appStoreSettingsRepo "subsnotifpro-go/internal/appstore/settings/repository"
	appStoreSettingsService "subsnotifpro-go/internal/appstore/settings/service"
	appStoreSubscriptionService "subsnotifpro-go/internal/appstore/subscription/service"
	appStoreUserService "subsnotifpro-go/internal/appstore/user/service"
	appStoreWebhookHandler "subsnotifpro-go/internal/appstore/webhooks/handler"
	appStoreWebhookService "subsnotifpro-go/internal/appstore/webhooks/service"
	"subsnotifpro-go/internal/constants"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	playstoreApiHandler "subsnotifpro-go/internal/playstore/api/handler"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreClientService "subsnotifpro-go/internal/playstore/client/service"
	dispatch "subsnotifpro-go/internal/playstore/dispatch"
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
	unifiedSubscriptionHandler "subsnotifpro-go/internal/subscription/handler"
	unifiedPublisher "subsnotifpro-go/internal/subscription/publisher"
	unifiedSubscriptionRepo "subsnotifpro-go/internal/subscription/repository"
	unifiedSubscriptionService "subsnotifpro-go/internal/subscription/service"
	userRepo "subsnotifpro-go/internal/users/repository"
	userService "subsnotifpro-go/internal/users/service"

	"sync"
	"syscall"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/health"
	"subsnotifpro-go/internal/migrations"
	queue "subsnotifpro-go/queue"

	"subsnotifpro-go/routes"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// // Initialize logger with options
	// logger := logger.New(
	// 	logger.WithLevel(logger.DebugLevel),
	// 	logger.WithCaller(true),
	// )

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize database and pass to repositories
	db, err := database.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.CloseDatabase(db)

	database.AutoMigrateTables(db, cfg)  // Auto-migrate tables
	migrations.ApplyCompositeIndexes(db) // Apply the composite indexes
	// Create materialized views
	migrations.CreateMaterializedViews(db)
	// Create view indexes
	migrations.CreateViewIndexes(db)

	// ✅ Initialize messaging backend based on configuration
	msgPublisher, err := messaging.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create message publisher: %v", err)
	}
	defer msgPublisher.Close()

	// ✅ Initialize health checker
	healthChecker := health.NewHealthChecker(db, string(cfg.MessagingType), "1.0.0")
	log.Println("✅ Health checker initialized")

	// Initialize RabbitMQ connection manager (only if using RabbitMQ)
	var rmqManager *queue.RabbitMQManager
	if cfg.MessagingType == config.MessagingTypeRabbitMQ {
		rmqManager = queue.NewRabbitMQManager(ctx, cfg.RabbitMQ)
		defer rmqManager.Close()

		// Get RabbitMQ channel
		ch, err := rmqManager.GetChannel()
		if err != nil {
			log.Fatalf("❌ Failed to get RabbitMQ channel: %v", err)
		}
		defer ch.Close()

		// Initialize RabbitMQ infrastructure
		if err := rmqManager.InitializeRabbitMQ(ctx, ch); err != nil {
			log.Fatalf("❌ Failed to initialize RabbitMQ: %v", err)
		}

		// Set RabbitMQ connection for health checks
		healthChecker.SetRabbitMQConnection(rmqManager.GetConnection())
	}

	// Initialize Service Bus client for health checks if using Service Bus
	if cfg.MessagingType == config.MessagingTypeServiceBus {
		// This is just for health checks - the actual publisher is created separately
		if cfg.ServiceBus.ConnectionString != "" {
			// We'll create a Service Bus client for health checks
			// Note: This is simplified - in production you might want to share the client
			log.Printf("📦 Service Bus configured for health checks")
		}
	}

	// ✅ Initialize Repositories, Services and Handlers

	log.Println("🔧  Initializing Playstore Services...")

	psClientService := playstoreClientService.NewPlaystoreClientService()
	psApiService := playstoreApiService.NewPlaystoreApiService(psClientService)

	psRtdnRepo := playstoreRTDNRepo.NewRTDNRepository(db)
	psSettingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository(db)
	psSubscriptionCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, getSyncBatchSize())

	// 🟢 Services

	// Create unified event publisher
	unifiedPublisher := unifiedPublisher.NewUnifiedEventPublisher(unifiedPublisher.UnifiedPublisherOpts{
		Publisher:  msgPublisher,
		Exchange:   cfg.RabbitMQ.UnifiedSubs.Exchange,
		RoutingKey: cfg.RabbitMQ.UnifiedSubs.RoutingKey,
		MaxRetries: cfg.RabbitMQ.MaxRetries,
		RetryDelay: cfg.RabbitMQ.RetryDelay,
	})

	dashboardRepo := unifiedSubscriptionRepo.NewDashboardRepository(db)
	dashboardSvc := unifiedSubscriptionService.NewDashboardService(dashboardRepo, 15*time.Minute)
	unifiedSubscriptionRepo := unifiedSubscriptionRepo.NewSubscriptionRepository(db)
	unifiedSubscriptionService := unifiedSubscriptionService.NewUnifiedSubscriptionService(db, unifiedPublisher, dashboardSvc, unifiedSubscriptionRepo)

	// Create domain-specific publishers
	googlePlayPublisher := dispatch.NewGooglePlayPublisher(msgPublisher, cfg)

	// ✅ Create a temporary placeholder for settingsService (declare first)

	userRepo := userRepo.NewUserRepository()
	userService := userService.NewUserService(userRepo)

	psUserRepo := playstoreUserRepo.NewPlaystoreUserRepository()

	psUserService := playstoreUserService.NewPlaystoreUserService(userService, psUserRepo)

	psCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, 50)

	psCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psCatalogRepo, psApiService)

	subscriptionRepo := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository(db)

	subscriptionService := playstoreSubscriptionService.NewPlaystoreSubscriptionService(db, subscriptionRepo, psUserService, psApiService, psCatalogService, unifiedSubscriptionService)

	psRtdnService := playstoreRTDNService.NewRTDNService(ctx, psRtdnRepo, psApiService, subscriptionService, db, googlePlayPublisher, rmqManager, &cfg.RabbitMQ)
	psSettingsService := playstoreSettingServicePkg.NewPlaystoreSettingsService(psSettingsRepo, psApiService, db)
	psClientService.SetAccountProvider(psSettingsService) // this works via interface

	// ✅ Step 4: Use the same instance
	psRtdnHandler := playstoreRtdnHandler.NewRTDNHandler(psRtdnService)
	psSettingsHandler := playstoreSettingsHandler.NewPlaystoreSettingsHandler(psSettingsService)
	psApiHandler := playstoreApiHandler.NewPlaystoreClientHandler(psApiService)
	psSubscriptionCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psSubscriptionCatalogRepo, psApiService)
	psSubscriptionCatalogHandler := playstoreCatalogHandler.NewSubscriptionCatalogHandler(psSubscriptionCatalogService)
	log.Println(" ✅ Initialized Playstore Services...")

	log.Println("🔧  Initializing Appstore Services...")

	appStorePublisher := appStoreDistpatch.NewAppStorePublisher(msgPublisher, cfg)
	appStoreUserService := appStoreUserService.NewAppStoreUserService(userService)
	appStoreSubscriptionService := appStoreSubscriptionService.NewAppStoreSubscriptionService(db, appStoreUserService, unifiedSubscriptionService)
	appStoreWebhookService := appStoreWebhookService.NewAppStoreNotificationsService(db, appStorePublisher, appStoreSubscriptionService)

	appStoreWebhookHandler := appStoreWebhookHandler.NewAppStoreNotificationsHandler(appStoreWebhookService)

	appStoreSettingsRepo := appStoreSettingsRepo.NewAppStoreSettingsRepository(db)
	appStoreSettingsService := appStoreSettingsService.NewAppStoreSettingsService(appStoreSettingsRepo)
	appStoreSettingsHandler := appStoreSettingsHandler.NewAppStoreSettingsHandler(appStoreSettingsService)

	dashBoardHandler := unifiedSubscriptionHandler.NewDashboardHandler(dashboardSvc)
	unifiedSubscriptionHandler := unifiedSubscriptionHandler.NewHandler(unifiedSubscriptionService)

	log.Println(" ✅ Initialized Appstore Services...")

	log.Println(" 🔌 Initializing Handlers...")
	deps := &routes.RouteDependencies{
		PlaystoreRTDNHandler:                psRtdnHandler,
		PlaystoreSettingsHandler:            psSettingsHandler,
		PlaystoreApiHandler:                 psApiHandler,
		PlaystoreSubscriptionCatalogHandler: psSubscriptionCatalogHandler,

		AppStoreWebhookHandler:      appStoreWebhookHandler,
		AppStoreSettingsHandler:     appStoreSettingsHandler,
		DashboardHandler:            dashBoardHandler,
		UnifiedSubscriptionsHandler: unifiedSubscriptionHandler,
		HealthChecker:               healthChecker,
	}

	router := routes.SetupRouter(deps)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}
	log.Println(" ✅ Initialized Handlers...")

	// ✅ Start HTTP Server (Non-Blocking)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🚀 Starting server on port %s...\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Add this right after starting the server
	log.Println("✅ Server startup complete - All systems operational")
	log.Println("====================================================")
	log.Printf("🔗 HTTP server listening on :%s", cfg.ServerPort)

	// Log messaging backend info
	switch cfg.MessagingType {
	case config.MessagingTypeRabbitMQ:
		if rmqManager != nil {
			log.Printf("📦 RabbitMQ connected: %s", rmqManager.SanitizeRabbitMQURL())
		}
	case config.MessagingTypeServiceBus:
		log.Printf("📦 Azure Service Bus connected")
	default:
		log.Printf("📦 Unknown messaging type: %s", cfg.MessagingType)
	}

	log.Println("====================================================")

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
