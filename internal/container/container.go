package container

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	appStoreDispatch "subsnotifpro-go/internal/appstore/dispatch"
	appStoreSettingsHandler "subsnotifpro-go/internal/appstore/settings/handler"
	appStoreSettingsRepo "subsnotifpro-go/internal/appstore/settings/repository"
	appStoreSettingsService "subsnotifpro-go/internal/appstore/settings/service"
	appStoreSubscriptionService "subsnotifpro-go/internal/appstore/subscription/service"
	appStoreUserService "subsnotifpro-go/internal/appstore/user/service"
	appStoreWebhookHandler "subsnotifpro-go/internal/appstore/webhooks/handler"
	appStoreWebhookService "subsnotifpro-go/internal/appstore/webhooks/service"
	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/health"
	middlewarePackage "subsnotifpro-go/internal/middleware"
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
	"subsnotifpro-go/queue"
	"subsnotifpro-go/routes"

	"subsnotifpro-go/config"
	"subsnotifpro-go/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Container manages all application dependencies with proper lifecycle management
type Container struct {
	// Core infrastructure
	Config    *config.Config
	DB        *gorm.DB
	Logger    *logrus.Logger
	Publisher messaging.MessagePublisher

	// Infrastructure services
	HealthChecker *health.HealthChecker
	RMQManager    *queue.RabbitMQManager
	SyncBatchSize int

	// Domain services - Playstore
	PlaystoreClientService              playstoreClientService.PlaystoreClientService
	PlaystoreApiService                 playstoreApiService.PlaystoreApiService
	PlaystoreRTDNService                playstoreRTDNService.RTDNService
	PlaystoreSettingsService            playstoreSettingServicePkg.PlaystoreSettingsService
	PlaystoreSubscriptionService        playstoreSubscriptionService.PlaystoreSubscriptionService
	PlaystoreSubscriptionCatalogService playstoreCatalogService.SubscriptionCatalogService
	PlaystoreUserService                playstoreUserService.PlaystoreUserService

	// Domain services - AppStore
	AppStoreUserService         appStoreUserService.AppStoreUserService
	AppStoreSubscriptionService appStoreSubscriptionService.AppStoreSubscriptionService
	AppStoreWebhookService      appStoreWebhookService.AppStoreNotificationsService
	AppStoreSettingsService     appStoreSettingsService.AppStoreSettingsService

	// Unified services
	UnifiedSubscriptionService unifiedSubscriptionService.UnifiedSubscriptionService
	DashboardService           unifiedSubscriptionService.DashboardService
	UserService                userService.UserService

	// Handlers
	PlaystoreHandlers *PlaystoreHandlers
	AppStoreHandlers  *AppStoreHandlers
	UnifiedHandlers   *UnifiedHandlers

	// Publishers
	GooglePlayPublisher *dispatch.GooglePlayPublisher
	AppStorePublisher   *appStoreDispatch.AppStorePublisher
	UnifiedPublisher    *unifiedPublisher.UnifiedEventPublisher

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// PlaystoreHandlers groups all Playstore-related handlers
type PlaystoreHandlers struct {
	RTDNHandler                *playstoreRtdnHandler.RTDNHandler
	SettingsHandler            *playstoreSettingsHandler.PlaystoreSettingsHandler
	ApiHandler                 *playstoreApiHandler.PlaystoreApiHandler
	EnhancedApiHandler         *playstoreApiHandler.EnhancedPlaystoreApiHandler
	SubscriptionCatalogHandler *playstoreCatalogHandler.SubscriptionCatalogHandler
}

// AppStoreHandlers groups all AppStore-related handlers
type AppStoreHandlers struct {
	WebhookHandler  *appStoreWebhookHandler.AppStoreNotificationsHandler
	SettingsHandler *appStoreSettingsHandler.AppStoreSettingsHandler
}

// UnifiedHandlers groups unified service handlers
type UnifiedHandlers struct {
	DashboardHandler    *unifiedSubscriptionHandler.DashboardHandler
	SubscriptionHandler *unifiedSubscriptionHandler.UnifiedSubscriptionsHandler
}

// NewContainer creates and initializes all application dependencies
func NewContainer(cfg *config.Config) (*Container, error) {
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	container := &Container{
		Config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize infrastructure dependencies
	if err := container.initializeInfrastructure(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize domain services
	if err := container.initializeDomainServices(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize domain services: %w", err)
	}

	// Initialize handlers
	if err := container.initializeHandlers(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	log.Println("✅ Container initialized successfully")
	return container, nil
}

// initializeInfrastructure sets up core infrastructure dependencies
func (c *Container) initializeInfrastructure() error {
	// Initialize logger
	c.Logger = logrus.New()
	c.Logger.SetLevel(logrus.InfoLevel)
	c.Logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize database
	db, err := database.ConnectDatabase(c.Config)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	c.DB = db

	// Apply migrations
	database.AutoMigrateTables(db, c.Config)
	log.Println("✅ Database migrations applied")

	// Initialize messaging
	publisher, err := messaging.NewPublisher(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create message publisher: %w", err)
	}
	c.Publisher = publisher

	// Initialize RabbitMQ manager if needed
	if c.Config.MessagingType == config.MessagingTypeRabbitMQ {
		c.RMQManager = queue.NewRabbitMQManager(c.ctx, c.Config.RabbitMQ)

		// Get channel and initialize
		ch, err := c.RMQManager.GetChannel()
		if err != nil {
			return fmt.Errorf("failed to get RabbitMQ channel: %w", err)
		}
		defer ch.Close()

		if err := c.RMQManager.InitializeRabbitMQ(c.ctx, ch); err != nil {
			return fmt.Errorf("failed to initialize RabbitMQ: %w", err)
		}
		log.Println("✅ RabbitMQ initialized")
	}

	// Initialize health checker
	c.HealthChecker = health.NewHealthChecker(db, string(c.Config.MessagingType), "1.0.0")
	if c.RMQManager != nil {
		c.HealthChecker.SetRabbitMQConnection(c.RMQManager.GetConnection())
	}

	// Set sync batch size
	c.SyncBatchSize = getSyncBatchSize()

	log.Println("✅ Infrastructure initialized")
	return nil
}

// initializeDomainServices sets up all domain services
func (c *Container) initializeDomainServices() error {
	// Initialize base services
	c.initializeBaseServices()

	// Initialize Playstore services
	if err := c.initializePlaystoreServices(); err != nil {
		return fmt.Errorf("failed to initialize Playstore services: %w", err)
	}

	// Initialize AppStore services
	if err := c.initializeAppStoreServices(); err != nil {
		return fmt.Errorf("failed to initialize AppStore services: %w", err)
	}

	// Initialize unified services
	if err := c.initializeUnifiedServices(); err != nil {
		return fmt.Errorf("failed to initialize unified services: %w", err)
	}

	log.Println("✅ Domain services initialized")
	return nil
}

// initializeBaseServices sets up foundational services
func (c *Container) initializeBaseServices() {
	// User services
	userRepo := userRepo.NewUserRepository()
	c.UserService = userService.NewUserService(userRepo)
}

// initializePlaystoreServices sets up all Playstore-related services
func (c *Container) initializePlaystoreServices() error {
	// Client and API services
	c.PlaystoreClientService = playstoreClientService.NewPlaystoreClientService()
	c.PlaystoreApiService = playstoreApiService.NewPlaystoreApiService(c.PlaystoreClientService)

	// Repositories
	psRtdnRepo := playstoreRTDNRepo.NewRTDNRepository(c.DB)
	psSettingsRepo := playstoreSettingRepo.NewPlaystoreSettingsRepository(c.DB)
	psUserRepo := playstoreUserRepo.NewPlaystoreUserRepository()
	psCatalogRepo := playstoreCatalogRepo.NewSubscriptionCatalogRepository(c.DB, 50)

	// Publishers
	c.GooglePlayPublisher = dispatch.NewGooglePlayPublisher(c.Publisher, c.Config)

	// Services
	c.PlaystoreUserService = playstoreUserService.NewPlaystoreUserService(c.UserService, psUserRepo)
	c.PlaystoreSubscriptionCatalogService = playstoreCatalogService.NewSubscriptionCatalogService(c.ctx, psCatalogRepo, c.PlaystoreApiService)
	c.PlaystoreSettingsService = playstoreSettingServicePkg.NewPlaystoreSettingsService(psSettingsRepo, c.PlaystoreApiService, c.DB)

	// Set account provider for client service
	c.PlaystoreClientService.SetAccountProvider(c.PlaystoreSettingsService)

	// Initialize subscription service (needs unified service, so we'll initialize it later)
	// We'll complete this in initializeUnifiedServices

	// RTDN service
	c.PlaystoreRTDNService = playstoreRTDNService.NewRTDNService(
		c.ctx,
		psRtdnRepo,
		c.PlaystoreApiService,
		nil, // Will be set later when subscription service is ready
		c.DB,
		c.GooglePlayPublisher,
		c.RMQManager,
		&c.Config.RabbitMQ,
	)

	log.Println("✅ Playstore services initialized")
	return nil
}

// initializeAppStoreServices sets up all AppStore-related services
func (c *Container) initializeAppStoreServices() error {
	// Publishers
	c.AppStorePublisher = appStoreDispatch.NewAppStorePublisher(c.Publisher, c.Config)

	// Services
	c.AppStoreUserService = appStoreUserService.NewAppStoreUserService(c.UserService)

	// Note: AppStore subscription service needs unified service, will be initialized later

	// Repositories and remaining services
	appStoreSettingsRepo := appStoreSettingsRepo.NewAppStoreSettingsRepository(c.DB)
	c.AppStoreSettingsService = appStoreSettingsService.NewAppStoreSettingsService(appStoreSettingsRepo)

	log.Println("✅ AppStore services initialized")
	return nil
}

// initializeUnifiedServices sets up unified services
func (c *Container) initializeUnifiedServices() error {
	// Unified publisher
	c.UnifiedPublisher = unifiedPublisher.NewUnifiedEventPublisher(unifiedPublisher.UnifiedPublisherOpts{
		Publisher:  c.Publisher,
		Exchange:   c.Config.RabbitMQ.UnifiedSubs.Exchange,
		RoutingKey: c.Config.RabbitMQ.UnifiedSubs.RoutingKey,
		MaxRetries: c.Config.RabbitMQ.MaxRetries,
		RetryDelay: c.Config.RabbitMQ.RetryDelay,
	})

	// Dashboard services
	dashboardRepo := unifiedSubscriptionRepo.NewDashboardRepository(c.DB)
	c.DashboardService = unifiedSubscriptionService.NewDashboardService(dashboardRepo, 15*time.Minute)

	// Unified subscription services
	unifiedSubscriptionRepo := unifiedSubscriptionRepo.NewSubscriptionRepository(c.DB)
	c.UnifiedSubscriptionService = unifiedSubscriptionService.NewUnifiedSubscriptionService(
		c.DB,
		c.UnifiedPublisher,
		c.DashboardService,
		unifiedSubscriptionRepo,
	)

	// Now complete Playstore subscription service
	subscriptionRepo := playstoreSubscriptionRepository.NewPlaystoreSubscriptionRepository(c.DB)
	c.PlaystoreSubscriptionService = playstoreSubscriptionService.NewPlaystoreSubscriptionService(
		c.DB,
		subscriptionRepo,
		c.PlaystoreUserService,
		c.PlaystoreApiService,
		c.PlaystoreSubscriptionCatalogService,
		c.UnifiedSubscriptionService,
	)

	// Complete AppStore subscription service
	c.AppStoreSubscriptionService = appStoreSubscriptionService.NewAppStoreSubscriptionService(
		c.DB,
		c.AppStoreUserService,
		c.UnifiedSubscriptionService,
	)

	// AppStore webhook service
	c.AppStoreWebhookService = appStoreWebhookService.NewAppStoreNotificationsService(
		c.DB,
		c.AppStorePublisher,
		c.AppStoreSubscriptionService,
	)

	log.Println("✅ Unified services initialized")
	return nil
}

// initializeHandlers sets up all HTTP handlers
func (c *Container) initializeHandlers() error {
	// Playstore handlers
	c.PlaystoreHandlers = &PlaystoreHandlers{
		RTDNHandler:                c.createPlaystoreRTDNHandler(),
		SettingsHandler:            c.createPlaystoreSettingsHandler(),
		ApiHandler:                 c.createPlaystoreApiHandler(),
		EnhancedApiHandler:         c.createEnhancedPlaystoreApiHandler(),
		SubscriptionCatalogHandler: c.createPlaystoreSubscriptionCatalogHandler(),
	}

	// AppStore handlers
	c.AppStoreHandlers = &AppStoreHandlers{
		WebhookHandler:  c.createAppStoreWebhookHandler(),
		SettingsHandler: c.createAppStoreSettingsHandler(),
	}

	// Unified handlers
	c.UnifiedHandlers = &UnifiedHandlers{
		DashboardHandler:    c.createDashboardHandler(),
		SubscriptionHandler: c.createUnifiedSubscriptionHandler(),
	}

	log.Println("✅ Handlers initialized")
	return nil
}

// Handler creation methods
func (c *Container) createPlaystoreRTDNHandler() *playstoreRtdnHandler.RTDNHandler {
	return playstoreRtdnHandler.NewRTDNHandler(c.PlaystoreRTDNService)
}

func (c *Container) createPlaystoreSettingsHandler() *playstoreSettingsHandler.PlaystoreSettingsHandler {
	return playstoreSettingsHandler.NewPlaystoreSettingsHandler(c.PlaystoreSettingsService)
}

func (c *Container) createPlaystoreApiHandler() *playstoreApiHandler.PlaystoreApiHandler {
	return playstoreApiHandler.NewPlaystoreClientHandler(c.PlaystoreApiService)
}

func (c *Container) createEnhancedPlaystoreApiHandler() *playstoreApiHandler.EnhancedPlaystoreApiHandler {
	return playstoreApiHandler.NewEnhancedPlaystoreApiHandler(c.PlaystoreApiService)
}

func (c *Container) createEnhancedMiddleware() *middlewarePackage.EnhancedMiddleware {
	return middlewarePackage.NewEnhancedMiddleware()
}

func (c *Container) createPlaystoreSubscriptionCatalogHandler() *playstoreCatalogHandler.SubscriptionCatalogHandler {
	return playstoreCatalogHandler.NewSubscriptionCatalogHandler(c.PlaystoreSubscriptionCatalogService)
}

func (c *Container) createAppStoreWebhookHandler() *appStoreWebhookHandler.AppStoreNotificationsHandler {
	return appStoreWebhookHandler.NewAppStoreNotificationsHandler(c.AppStoreWebhookService)
}

func (c *Container) createAppStoreSettingsHandler() *appStoreSettingsHandler.AppStoreSettingsHandler {
	return appStoreSettingsHandler.NewAppStoreSettingsHandler(c.AppStoreSettingsService)
}

func (c *Container) createDashboardHandler() *unifiedSubscriptionHandler.DashboardHandler {
	return unifiedSubscriptionHandler.NewDashboardHandler(c.DashboardService)
}

func (c *Container) createUnifiedSubscriptionHandler() *unifiedSubscriptionHandler.UnifiedSubscriptionsHandler {
	return unifiedSubscriptionHandler.NewHandler(c.UnifiedSubscriptionService)
}

// GetRouteDependencies creates the route dependencies struct
func (c *Container) GetRouteDependencies() *routes.RouteDependencies {
	return &routes.RouteDependencies{
		PlaystoreRTDNHandler:                c.PlaystoreHandlers.RTDNHandler,
		PlaystoreSettingsHandler:            c.PlaystoreHandlers.SettingsHandler,
		PlaystoreApiHandler:                 c.PlaystoreHandlers.ApiHandler,
		EnhancedPlaystoreApiHandler:         c.PlaystoreHandlers.EnhancedApiHandler,
		PlaystoreSubscriptionCatalogHandler: c.PlaystoreHandlers.SubscriptionCatalogHandler,
		AppStoreWebhookHandler:              c.AppStoreHandlers.WebhookHandler,
		AppStoreSettingsHandler:             c.AppStoreHandlers.SettingsHandler,
		DashboardHandler:                    c.UnifiedHandlers.DashboardHandler,
		UnifiedSubscriptionsHandler:         c.UnifiedHandlers.SubscriptionHandler,
		HealthChecker:                       c.HealthChecker,
		EnhancedMiddleware:                  c.createEnhancedMiddleware(),
	}
}

// Close gracefully shuts down all dependencies
func (c *Container) Close() error {
	log.Println("🚦 Starting container shutdown...")

	// Cancel context to stop background operations
	if c.cancel != nil {
		c.cancel()
	}

	// Close messaging publisher
	if c.Publisher != nil {
		if err := c.Publisher.Close(); err != nil {
			log.Printf("❌ Error closing publisher: %v", err)
		}
	}

	// Close RabbitMQ manager
	if c.RMQManager != nil {
		c.RMQManager.Close()
	}

	// Close database
	if c.DB != nil {
		database.CloseDatabase(c.DB)
	}

	log.Println("✅ Container shutdown complete")
	return nil
}

// Helper function to get sync batch size from environment
func getSyncBatchSize() int {
	if envSize := os.Getenv("SYNC_BATCH_SIZE"); envSize != "" {
		if parsedSize, err := strconv.Atoi(envSize); err == nil {
			return parsedSize
		}
	}
	return constants.DEFAULT_SYNC_BATCH_SIZE
}
