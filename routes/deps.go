package routes

import (
	"github.com/streadway/amqp"

	playstoreApiHandler "subsnotifpro-go/internal/playstore/api/handler"
	playstoreSubscriptionCatalogHandler "subsnotifpro-go/internal/playstore/products/handler"
	playstoreRTDNHandler "subsnotifpro-go/internal/playstore/rtdn/handler"
	playstoreSettingsHandler "subsnotifpro-go/internal/playstore/settings/handler"
)

type RouteDependencies struct {
	RabbitMQChannel *amqp.Channel

	PlaystoreRTDNHandler                *playstoreRTDNHandler.RTDNHandler
	PlaystoreSettingsHandler            *playstoreSettingsHandler.PlaystoreSettingsHandler
	PlaystoreApiHandler                 *playstoreApiHandler.PlaystoreApiHandler
	PlaystoreSubscriptionCatalogHandler *playstoreSubscriptionCatalogHandler.SubscriptionCatalogHandler
}
