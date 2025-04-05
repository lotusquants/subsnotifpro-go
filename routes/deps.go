package routes

import (
	playstoreApiHandler "subsnotifpro-go/internal/playstore/api/handler"
	playstoreSubscriptionCatalogHandler "subsnotifpro-go/internal/playstore/products/handler"
	playstoreRTDNHandler "subsnotifpro-go/internal/playstore/rtdn/handler"
	playstoreSettingsHandler "subsnotifpro-go/internal/playstore/settings/handler"
)

type RouteDependencies struct {
	PlaystoreRTDNHandler                *playstoreRTDNHandler.RTDNHandler
	PlaystoreSettingsHandler            *playstoreSettingsHandler.PlaystoreSettingsHandler
	PlaystoreApiHandler                 *playstoreApiHandler.PlaystoreApiHandler
	PlaystoreSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler
}
