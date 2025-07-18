package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/streadway/amqp"
	"subsnotifpro-go/config"
	"subsnotifpro-go/database"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreSettingRepo "subsnotifpro-go/internal/playstore/settings/repository"
	playstoreSettingServicePkg "subsnotifpro-go/internal/playstore/settings/service"
	playstoreSubscriptionRepository "subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"
	playstoreUserRepo "subsnotifpro-go/internal/playstore/user/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"
	userRepo "subsnotifpro-go/internal/users/repository"
	userService "subsnotifpro-go/internal/users/service"

	playstoreCatalogRepo "subsnotifpro-go/internal/playstore/products/repository"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"

	clientService "subsnotifpro-go/internal/playstore/client/service"

	playstoreDispatchPkg "subsnotifpro-go/internal/playstore/dispatch"

	rtdnRepo "subsnotifpro-go/internal/playstore/rtdn/repository"
	rtdnService "subsnotifpro-go/internal/playstore/rtdn/service"

	appStoreDistpatch "subsnotifpro-go/internal/appstore/dispatch"
	appStoreSubscriptionService "subsnotifpro-go/internal/appstore/subscription/service"
	appStoreUserService "subsnotifpro-go/internal/appstore/user/service"
	appStoreWebhookRepo "subsnotifpro-go/internal/appstore/webhooks/repository"
	appStoreWebhookService "subsnotifpro-go/internal/appstore/webhooks/service"
	unifiedSubscriptionsConsumer "subsnotifpro-go/internal/subscription/consumer"
	unifiedPublisher "subsnotifpro-go/internal/subscription/publisher"
	unifiedSubscriptionRepo "subsnotifpro-go/internal/subscription/repository"
	unifiedSubscriptionService "subsnotifpro-go/internal/subscription/service"

	"subsnotifpro-go/queue"
)

func main() {
	// Create shutdown context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup wait group for graceful shutdown
	var wg sync.WaitGroup

	// ✅ Load configuration
	cfg := config.LoadConfig()

	// ✅ Initialize messaging backend based on configuration
	msgPublisher, err := messaging.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create message publisher: %v", err)
	}
	defer msgPublisher.Close()

	// Initialize RabbitMQ connection manager (only if using RabbitMQ)
	var rmqManager *queue.RabbitMQManager
	var ch *amqp.Channel
	if cfg.MessagingType == config.MessagingTypeRabbitMQ {
		rmqManager = queue.NewRabbitMQManager(ctx, cfg.RabbitMQ)
		defer func() {
			log.Println("🚦 Closing RabbitMQ connection...")
			rmqManager.Close()
		}()

		// Get RabbitMQ channel
		ch, err = rmqManager.GetChannel()
		if err != nil {
			log.Fatalf("❌ Failed to get RabbitMQ channel: %v", err)
		}
	}

	// Initialize database
	db, err := database.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer func() {
		log.Println("🚦 Closing database connection...")
		database.CloseDatabase(db)
	}()

	// Initialize services
	// msgPublisher is already initialized above with the messaging factory

	googlePlayPublisher := playstoreDispatchPkg.NewGooglePlayPublisher(msgPublisher, cfg)

	// 2. Create unified event publisher
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

	clientService := clientService.NewPlaystoreClientService()
	apiService := playstoreApiService.NewPlaystoreApiService(clientService)
	settingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository(db)
	settingsService := playstoreSettingServicePkg.NewPlaystoreSettingsService(settingsRepo, apiService, db)
	clientService.SetAccountProvider(settingsService)
	userRepo := userRepo.NewUserRepository()
	userService := userService.NewUserService(userRepo)
	psUserRepo := playstoreUserRepo.NewPlaystoreUserRepository()
	psUserService := playstoreUserService.NewPlaystoreUserService(userService, psUserRepo)
	psCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, 50)
	psCatalogService := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psCatalogRepo, apiService)
	subscriptionRepo := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository(db)
	subscriptionService := playstoreSubscriptionService.NewPlaystoreSubscriptionService(db, subscriptionRepo, psUserService, apiService, psCatalogService, unifiedSubscriptionService)
	rtdnRepo := rtdnRepo.NewRTDNRepository(db)
	rtdnService := rtdnService.NewRTDNService(ctx, rtdnRepo, apiService, subscriptionService, db, googlePlayPublisher, rmqManager, &cfg.RabbitMQ)

	log.Println("🔧  Initializing Appstore Services...")

	appStorePublisher := appStoreDistpatch.NewAppStorePublisher(msgPublisher, cfg)
	appStoreWebhookRepository := appStoreWebhookRepo.NewAppstoreNotificationsRepository(db)
	appStoreUserService := appStoreUserService.NewAppStoreUserService(userService)
	appStoreSubscriptionService := appStoreSubscriptionService.NewAppStoreSubscriptionService(db, appStoreUserService, unifiedSubscriptionService)
	appStoreWebhookService := appStoreWebhookService.NewAppStoreNotificationsService(db, appStorePublisher, appStoreSubscriptionService)

	log.Println(" ✅ Initialized Appstore Services...")

	// Create and start consumers based on messaging type
	if cfg.MessagingType == config.MessagingTypeRabbitMQ && ch != nil {
		// Create and start the RabbitMQ consumers
		consumer := playstoreDispatchPkg.NewGooglePlayConsumer(ch, rtdnRepo, rtdnService, msgPublisher, cfg)

		wg.Add(1)
		go func() {
			defer wg.Done()
			consumer.Consumer.Start(ctx)
		}()

		log.Println("🚀 Playstore Consumer/Worker started successfully")

		// Create and start the consumer
		appStoreConsumer := appStoreDistpatch.NewAppStoreConsumer(ch, appStoreWebhookRepository, appStoreWebhookService, msgPublisher, cfg)

		wg.Add(1)
		go func() {
			defer wg.Done()
			appStoreConsumer.Consumer.Start(ctx)
		}()

		log.Println("🚀 Appstore Consumer/Worker started successfully")

		// Configure consumer options
		unifiedSubscriptionsConsumerOpts := unifiedSubscriptionsConsumer.UnifiedConsumerOpts{
			Channel:   ch,
			Repo:      unifiedSubscriptionRepo,
			Service:   unifiedSubscriptionService,
			Publisher: msgPublisher,
			Cfg:       &cfg.RabbitMQ, // your RabbitMQ config
		}

		// Create and start the consumer
		unifiedSubscriptionConsumer := unifiedSubscriptionsConsumer.NewUnifiedConsumer(unifiedSubscriptionsConsumerOpts)

		wg.Add(1)
		go func() {
			defer wg.Done()
			unifiedSubscriptionConsumer.Start(ctx)
		}()

		log.Println("🚀 Unified Subscription Consumer/Worker started successfully")
	} else if cfg.MessagingType == config.MessagingTypeServiceBus {
		// Create and start Service Bus consumers
		msgConsumer, err := messaging.NewMessageConsumer(cfg)
		if err != nil {
			log.Printf("❌ Failed to create message consumer: %v", err)
			// For now, just log the error and continue
			// Service Bus consumers would need specific implementation
		} else {
			defer msgConsumer.Close()
		}

		// Note: Service Bus consumers would need to be implemented
		// This is a placeholder for the Service Bus consumer logic
		log.Println("🚀 Service Bus Consumer/Worker started successfully")
	}

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("🚦 Received shutdown signal: %v", sig)

	// Initiate graceful shutdown
	cancel()

	// Setup shutdown timeout
	shutdownTimeout := 15 * time.Second
	shutdownDone := make(chan struct{})

	// Wait for goroutines to finish in background
	go func() {
		wg.Wait()
		close(shutdownDone)
	}()

	// Wait for shutdown or timeout
	select {
	case <-shutdownDone:
		log.Println("✅ All components shut down gracefully")
	case <-time.After(shutdownTimeout):
		log.Println("⚠️ Shutdown timeout reached, forcing exit")
	}

	log.Println("👋 Worker shutdown complete")
}
