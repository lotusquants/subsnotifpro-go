package routes

import (
	appStoreSettingsHandler "subsnotifpro-go/internal/appstore/settings/handler"
	appStoreWebhookHandler "subsnotifpro-go/internal/appstore/webhooks/handler"
	"subsnotifpro-go/internal/auth"
	"subsnotifpro-go/internal/health"
	playstoreApiHandler "subsnotifpro-go/internal/playstore/api/handler"
	playstoreSubscriptionCatalogHandler "subsnotifpro-go/internal/playstore/products/handler"
	playstoreRTDNHandler "subsnotifpro-go/internal/playstore/rtdn/handler"
	playstoreSettingsHandler "subsnotifpro-go/internal/playstore/settings/handler"
	unifiedSubscriptionHandler "subsnotifpro-go/internal/subscription/handler"
)

type RouteDependencies struct {
	PlaystoreRTDNHandler                *playstoreRTDNHandler.RTDNHandler
	PlaystoreSettingsHandler            *playstoreSettingsHandler.PlaystoreSettingsHandler
	PlaystoreApiHandler                 *playstoreApiHandler.PlaystoreApiHandler
	PlaystoreSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler
	AppStoreWebhookHandler              *appStoreWebhookHandler.AppStoreNotificationsHandler
	AppStoreSettingsHandler             *appStoreSettingsHandler.AppStoreSettingsHandler
	DashboardHandler                    *unifiedSubscriptionHandler.DashboardHandler
	UnifiedSubscriptionsHandler         *unifiedSubscriptionHandler.UnifiedSubscriptionsHandler
	HealthChecker                       *health.HealthChecker
	
	// Authentication components
	AuthHandler    *auth.AuthHandler
	AuthMiddleware *auth.AuthMiddleware
}
