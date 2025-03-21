package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	authHandler "subsnotifpro-go/internal/auth/handler"
	authRepo "subsnotifpro-go/internal/auth/repository"
	authService "subsnotifpro-go/internal/auth/service"
	"subsnotifpro-go/internal/constants"
	playstoreClientHandler "subsnotifpro-go/internal/google_playstore/client/handler"
	playstoreClientService "subsnotifpro-go/internal/google_playstore/client/service"
	playstoreRTDNHandler "subsnotifpro-go/internal/google_playstore/rtdn/handler"
	playstoreRTDNService "subsnotifpro-go/internal/google_playstore/rtdn/service"
	playstoreSettingsHandler "subsnotifpro-go/internal/google_playstore/settings/handler"
	playstoreSettingService "subsnotifpro-go/internal/google_playstore/settings/service"
	playstoreSubscriptionSyncHandler "subsnotifpro-go/internal/google_playstore/subscription_catalog/handler"
	playstoreSubscriptionSyncService "subsnotifpro-go/internal/google_playstore/subscription_catalog/service"
	"sync"
	"syscall"
	"time"

	playstoreRTDNRepo "subsnotifpro-go/internal/google_playstore/rtdn/repository"
	playstoreSettingRepo "subsnotifpro-go/internal/google_playstore/settings/repository"
	playstoreSubscriptionSyncRepo "subsnotifpro-go/internal/google_playstore/subscription_catalog/repository"

	tenantHandlerPkg "subsnotifpro-go/internal/tenant/handler"
	tenantRepo "subsnotifpro-go/internal/tenant/repository"
	tenantService "subsnotifpro-go/internal/tenant/service"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/database"
	"subsnotifpro-go/internal/messaging"
	"subsnotifpro-go/routes"
)

func main() {
	// ✅ Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize database and pass to repositories
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	database.AutoMigrateTables(db)     // Auto-migrate tables
	database.ApplyCompositeIndexes(db) // Apply the composite indexes

	// 🌱 Run Seeders
	if err := database.SeedDatabase(db); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

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

	// ✅ Initialize Repositories, Services and Handlers

	// Initialize RTDN repository
	psRtdnRepo := playstoreRTDNRepo.NewRTDNRepository(db)
	// Initialize playstore Settings repository
	psSettingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository()

	// Initialize  playstore client service
	psClientService := playstoreClientService.NewGooglePlayClientService(ctx, psSettingsRepo, db)

	// Initialize RTDN service
	psRtdnService := playstoreRTDNService.NewRTDNService(ctx, psRtdnRepo, psClientService)
	// Initialize RTDN handler
	psRtdnHandler := playstoreRTDNHandler.NewRTDNHandler(psRtdnService)

	// Initialize  playstore Settings service
	psSettingsService := playstoreSettingService.NewPlaystoreSettingsService(db, psSettingsRepo)
	// Initialize  playstore Settings handler
	psSettingsHandler := playstoreSettingsHandler.NewPlaystoreSettingsHandler(psSettingsService)

	// Initialize  playstore client handler
	psClientHandler := playstoreClientHandler.NewPlaystoreClientHandler(psClientService)

	syncBatchSize := constants.DEFAULT_SYNC_BATCH_SIZE
	if envSize := os.Getenv("SYNC_BATCH_SIZE"); envSize != "" {
		if parsedSize, err := strconv.Atoi(envSize); err == nil {
			syncBatchSize = parsedSize
		}
	}

	// ✅ Auth setup
	authRepository := authRepo.NewAdminRepository()
	authService := authService.NewAdminService(db, authRepository)
	authHandler := authHandler.NewAuthHandler(authService)

	// ✅ Tenant setup
	tenantRepository := tenantRepo.NewTenantRepository()
	tenantSvc := tenantService.NewTenantService(db, tenantRepository)
	tenantHandler := tenantHandlerPkg.NewTenantHandler(tenantSvc)

	appRepository := tenantRepo.NewAppRepository() // Assuming same package
	appSvc := tenantService.NewAppService(db, appRepository)
	appHandler := tenantHandlerPkg.NewAppHandler(appSvc)

	// Initialize Subscription Catalog repository, service, and handler
	// Initialize SubscriptionCatalog repository
	psSubscriptionCatalogRepo := playstoreSubscriptionSyncRepo.NewSubscriptionCatalogRepository(db, syncBatchSize)
	// Initialize SubscriptionCatalog service
	psSubscriptionCatalogService := playstoreSubscriptionSyncService.NewSubscriptionCatalogService(ctx, psSubscriptionCatalogRepo, psClientService)
	// Initialize SubscriptionCatalog handler
	psSubscriptionCatalogHandler := playstoreSubscriptionSyncHandler.NewSubscriptionCatalogHandler(psSubscriptionCatalogService)

	// ✅ Setup HTTP server and pass handlers to routes
	router := routes.SetupRouter(ch, psRtdnHandler, psSettingsHandler, psClientHandler, psSubscriptionCatalogHandler, authHandler, tenantHandler, appHandler)

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
	messaging.CloseRabbitMQ()

	log.Println("✅ Server shutdown complete")
}
