package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"subsnotifpro-go/cmd/worker/app"
	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/container"
	"subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/queue"

	// Service imports
	appStoreDispatch "subsnotifpro-go/internal/appstore/dispatch"
	appStoreSubscriptionService "subsnotifpro-go/internal/appstore/subscription/service"
	appStoreUserService "subsnotifpro-go/internal/appstore/user/service"
	appStoreWebhookRepo "subsnotifpro-go/internal/appstore/webhooks/repository"
	appStoreWebhookService "subsnotifpro-go/internal/appstore/webhooks/service"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	clientService "subsnotifpro-go/internal/playstore/client/service"
	playstoreDispatchPkg "subsnotifpro-go/internal/playstore/dispatch"
	playstoreCatalogRepo "subsnotifpro-go/internal/playstore/products/repository"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"
	rtdnRepo "subsnotifpro-go/internal/playstore/rtdn/repository"
	rtdnService "subsnotifpro-go/internal/playstore/rtdn/service"
	playstoreSettingRepo "subsnotifpro-go/internal/playstore/settings/repository"
	playstoreSettingServicePkg "subsnotifpro-go/internal/playstore/settings/service"
	playstoreSubscriptionRepository "subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"
	playstoreUserRepo "subsnotifpro-go/internal/playstore/user/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"
	unifiedSubscriptionsConsumer "subsnotifpro-go/internal/subscription/consumer"
	unifiedPublisher "subsnotifpro-go/internal/subscription/publisher"
	unifiedSubscriptionRepo "subsnotifpro-go/internal/subscription/repository"
	unifiedSubscriptionService "subsnotifpro-go/internal/subscription/service"
	userRepo "subsnotifpro-go/internal/users/repository"
	userService "subsnotifpro-go/internal/users/service"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
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

	// Initialize worker dependencies
	workerDeps, err := initializeWorkerDependencies(ctx, cfg, container)
	if err != nil {
		log.Fatalf("❌ Failed to initialize worker dependencies: %v", err)
	}

	// Create application with framework
	application, err := app.NewApplication(
		ctx,
		BuildInfo,
		app.WithDependencies(workerDeps),
		app.WithHealthCheck(func(ctx context.Context) error {
			// Add health checks for worker dependencies
			if workerDeps.MessagePublisher == nil {
				return fmt.Errorf("message publisher is not available")
			}
			return nil
		}),
		app.WithShutdownFunc(func() error {
			// Clean up worker dependencies
			if workerDeps.MessagePublisher != nil {
				workerDeps.MessagePublisher.Close()
			}
			if workerDeps.RabbitMQManager != nil {
				workerDeps.RabbitMQManager.Close()
			}
			return container.Close()
		}),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create worker application: %v", err)
	}

	// Initialize and run the application
	if err := application.Initialize(ctx); err != nil {
		log.Fatalf("❌ Failed to initialize worker application: %v", err)
	}

	// Run the application (blocks until shutdown)
	if err := application.Run(ctx); err != nil {
		log.Fatalf("❌ Worker application error: %v", err)
	}
}

// initializeWorkerDependencies sets up all the worker-specific dependencies
func initializeWorkerDependencies(ctx context.Context, cfg *config.Config, container *container.Container) (*app.WorkerDependencies, error) {
	log.Println("🔧 Initializing worker dependencies...")

	// Get database from container
	db := container.DB

	// Initialize messaging backend
	msgPublisher, err := messaging.NewPublisher(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create message publisher: %w", err)
	}

	// Initialize RabbitMQ connection manager (only if using RabbitMQ)
	var rmqManager *queue.RabbitMQManager
	var ch *amqp.Channel
	if cfg.MessagingType == config.MessagingTypeRabbitMQ {
		rmqManager = queue.NewRabbitMQManager(ctx, cfg.RabbitMQ)

		// Get RabbitMQ channel
		ch, err = rmqManager.GetChannel()
		if err != nil {
			return nil, fmt.Errorf("failed to get RabbitMQ channel: %w", err)
		}
	}

	// Initialize all services using the same pattern as the original main.go
	services, err := initializeServices(ctx, cfg, db, msgPublisher, rmqManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Create consumers based on messaging type
	var consumers []app.WorkerConsumer
	if cfg.MessagingType == config.MessagingTypeRabbitMQ && ch != nil {
		consumers = append(consumers,
			app.NewPlaystoreConsumerWrapper(ch, services.RTDNRepo, services.RTDNService, msgPublisher, cfg),
			app.NewAppStoreConsumerWrapper(ch, services.AppStoreWebhookRepo, services.AppStoreWebhookService, msgPublisher, cfg),
			app.NewUnifiedConsumerWrapper(unifiedSubscriptionsConsumer.UnifiedConsumerOpts{
				Channel:   ch,
				Repo:      services.UnifiedSubscriptionRepo,
				Service:   services.UnifiedSubscriptionService,
				Publisher: msgPublisher,
				Cfg:       &cfg.RabbitMQ,
			}),
		)
	} else if cfg.MessagingType == config.MessagingTypeServiceBus {
		// Service Bus consumers would be added here
		log.Println("🚀 Service Bus consumers not yet implemented")
	}

	deps := &app.WorkerDependencies{
		MessagePublisher: msgPublisher,
		RabbitMQManager:  rmqManager,
		Channel:          ch,
		Consumers:        consumers,
	}

	log.Printf("✅ Worker dependencies initialized with %d consumers", len(consumers))
	return deps, nil
}

// Services holds all the initialized services
type Services struct {
	// PlayStore services
	ClientService       clientService.PlaystoreClientService
	APIService          playstoreApiService.PlaystoreApiService
	SettingsRepo        playstoreSettingRepo.PlaystoreSettingsRepository
	SettingsService     playstoreSettingServicePkg.PlaystoreSettingsService
	UserRepo            userRepo.UserRepository
	UserService         userService.UserService
	PSUserRepo          playstoreUserRepo.PlaystoreUserRepository
	PSUserService       playstoreUserService.PlaystoreUserService
	PSCatalogRepo       playstoreCatalogRepo.SubscriptionCatalogRepository
	PSCatalogService    playstoreCatalogService.SubscriptionCatalogService
	SubscriptionRepo    playstoreSubscriptionRepository.PlaystoreSubscriptionRepository
	SubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService
	RTDNRepo            rtdnRepo.RTDNRepository
	RTDNService         rtdnService.RTDNService

	// AppStore services
	AppStoreWebhookRepo         appStoreWebhookRepo.AppstoreNotificationsRepository
	AppStoreUserService         appStoreUserService.AppStoreUserService
	AppStoreSubscriptionService appStoreSubscriptionService.AppStoreSubscriptionService
	AppStoreWebhookService      appStoreWebhookService.AppStoreNotificationsService

	// Unified services
	UnifiedSubscriptionRepo    unifiedSubscriptionRepo.SubscriptionRepository
	UnifiedSubscriptionService unifiedSubscriptionService.UnifiedSubscriptionService
}

// initializeServices initializes all the services using the same logic as the original main.go
func initializeServices(ctx context.Context, cfg *config.Config, db *gorm.DB, msgPublisher messaging.MessagePublisher, rmqManager *queue.RabbitMQManager) (*Services, error) {
	log.Println("🔧 Initializing services...")

	// Initialize publishers
	googlePlayPublisher := playstoreDispatchPkg.NewGooglePlayPublisher(msgPublisher, cfg)
	appStorePublisher := appStoreDispatch.NewAppStorePublisher(msgPublisher, cfg)

	// Create unified event publisher
	unifiedEventPublisher := unifiedPublisher.NewUnifiedEventPublisher(unifiedPublisher.UnifiedPublisherOpts{
		Publisher:  msgPublisher,
		Exchange:   cfg.RabbitMQ.UnifiedSubs.Exchange,
		RoutingKey: cfg.RabbitMQ.UnifiedSubs.RoutingKey,
		MaxRetries: cfg.RabbitMQ.MaxRetries,
		RetryDelay: cfg.RabbitMQ.RetryDelay,
	})

	// Initialize repositories and services in the same order as original
	dashboardRepo := unifiedSubscriptionRepo.NewDashboardRepository(db)
	dashboardSvc := unifiedSubscriptionService.NewDashboardService(dashboardRepo, 15*time.Minute)
	unifiedSubRepo := unifiedSubscriptionRepo.NewSubscriptionRepository(db)
	unifiedSubService := unifiedSubscriptionService.NewUnifiedSubscriptionService(db, unifiedEventPublisher, dashboardSvc, unifiedSubRepo)

	clientSvc := clientService.NewPlaystoreClientService()
	apiService := playstoreApiService.NewPlaystoreApiService(clientSvc)
	settingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository(db)
	settingsService := playstoreSettingServicePkg.NewPlaystoreSettingsService(settingsRepo, apiService, db)
	clientSvc.SetAccountProvider(settingsService)

	userRepository := userRepo.NewUserRepository()
	userSvc := userService.NewUserService(userRepository)
	psUserRepository := playstoreUserRepo.NewPlaystoreUserRepository()
	psUserSvc := playstoreUserService.NewPlaystoreUserService(userSvc, psUserRepository)
	psCatalogRepository := playstoreCatalogRepo.NewSubscriptionCatalogRepository(db, 50)
	psCatalogSvc := playstoreCatalogService.NewSubscriptionCatalogService(ctx, psCatalogRepository, apiService)
	subscriptionRepository := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository(db)
	subscriptionSvc := playstoreSubscriptionService.NewPlaystoreSubscriptionService(db, subscriptionRepository, psUserSvc, apiService, psCatalogSvc, unifiedSubService)
	rtdnRepository := rtdnRepo.NewRTDNRepository(db)
	rtdnSvc := rtdnService.NewRTDNService(ctx, rtdnRepository, apiService, subscriptionSvc, db, googlePlayPublisher, rmqManager, &cfg.RabbitMQ)

	// Initialize AppStore services
	appStoreWebhookRepository := appStoreWebhookRepo.NewAppstoreNotificationsRepository(db)
	appStoreUserSvc := appStoreUserService.NewAppStoreUserService(userSvc)
	appStoreSubscriptionSvc := appStoreSubscriptionService.NewAppStoreSubscriptionService(db, appStoreUserSvc, unifiedSubService)
	appStoreWebhookSvc := appStoreWebhookService.NewAppStoreNotificationsService(db, appStorePublisher, appStoreSubscriptionSvc)

	services := &Services{
		ClientService:               clientSvc,
		APIService:                  apiService,
		SettingsRepo:                settingsRepo,
		SettingsService:             settingsService,
		UserRepo:                    userRepository,
		UserService:                 userSvc,
		PSUserRepo:                  psUserRepository,
		PSUserService:               psUserSvc,
		PSCatalogRepo:               psCatalogRepository,
		PSCatalogService:            psCatalogSvc,
		SubscriptionRepo:            subscriptionRepository,
		SubscriptionService:         subscriptionSvc,
		RTDNRepo:                    rtdnRepository,
		RTDNService:                 rtdnSvc,
		AppStoreWebhookRepo:         appStoreWebhookRepository,
		AppStoreUserService:         appStoreUserSvc,
		AppStoreSubscriptionService: appStoreSubscriptionSvc,
		AppStoreWebhookService:      appStoreWebhookSvc,
		UnifiedSubscriptionRepo:     unifiedSubRepo,
		UnifiedSubscriptionService:  unifiedSubService,
	}

	log.Println("✅ All services initialized successfully")
	return services, nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
