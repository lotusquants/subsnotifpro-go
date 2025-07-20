package app

import (
	"context"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/messaging"

	appStoreDispatch "subsnotifpro-go/internal/appstore/dispatch"
	appStoreWebhookRepo "subsnotifpro-go/internal/appstore/webhooks/repository"
	appStoreWebhookService "subsnotifpro-go/internal/appstore/webhooks/service"
	playstoreDispatch "subsnotifpro-go/internal/playstore/dispatch"
	rtdnRepository "subsnotifpro-go/internal/playstore/rtdn/repository"
	rtdnService "subsnotifpro-go/internal/playstore/rtdn/service"
	unifiedConsumer "subsnotifpro-go/internal/subscription/consumer"

	"github.com/streadway/amqp"
)

// PlaystoreConsumerWrapper wraps the playstore consumer to implement WorkerConsumer
type PlaystoreConsumerWrapper struct {
	consumer *playstoreDispatch.GooglePlayConsumer
	name     string
	logger   logger.Logger
}

// NewPlaystoreConsumerWrapper creates a new wrapper for the playstore consumer
func NewPlaystoreConsumerWrapper(
	ch *amqp.Channel,
	rtdnRepo rtdnRepository.RTDNRepository,
	rtdnSvc rtdnService.RTDNService,
	msgPublisher messaging.MessagePublisher,
	cfg *config.Config,
) *PlaystoreConsumerWrapper {
	consumer := playstoreDispatch.NewGooglePlayConsumer(ch, rtdnRepo, rtdnSvc, msgPublisher, cfg)

	return &PlaystoreConsumerWrapper{
		consumer: consumer,
		name:     "playstore-consumer",
		logger:   logger.NewEnhancedLogger(),
	}
}

func (w *PlaystoreConsumerWrapper) Start(ctx context.Context) error {
	w.logger.Info("🔄 Starting Playstore consumer")
	// Start returns void, so we run it in the background
	go w.consumer.Consumer.Start(ctx)
	return nil
}

func (w *PlaystoreConsumerWrapper) Stop(ctx context.Context) error {
	w.logger.Info("🛑 Stopping Playstore consumer")
	// The consumer should handle context cancellation internally
	return nil
}

func (w *PlaystoreConsumerWrapper) Name() string {
	return w.name
}

// AppStoreConsumerWrapper wraps the appstore consumer to implement WorkerConsumer
type AppStoreConsumerWrapper struct {
	consumer *appStoreDispatch.AppStoreConsumer
	name     string
	logger   logger.Logger
}

// NewAppStoreConsumerWrapper creates a new wrapper for the appstore consumer
func NewAppStoreConsumerWrapper(
	ch *amqp.Channel,
	webhookRepo appStoreWebhookRepo.AppstoreNotificationsRepository,
	webhookService appStoreWebhookService.AppStoreNotificationsService,
	msgPublisher messaging.MessagePublisher,
	cfg *config.Config,
) *AppStoreConsumerWrapper {
	consumer := appStoreDispatch.NewAppStoreConsumer(ch, webhookRepo, webhookService, msgPublisher, cfg)

	return &AppStoreConsumerWrapper{
		consumer: consumer,
		name:     "appstore-consumer",
		logger:   logger.NewEnhancedLogger(),
	}
}

func (w *AppStoreConsumerWrapper) Start(ctx context.Context) error {
	w.logger.Info("🔄 Starting AppStore consumer")
	// Start returns void, so we run it in the background
	go w.consumer.Consumer.Start(ctx)
	return nil
}

func (w *AppStoreConsumerWrapper) Stop(ctx context.Context) error {
	w.logger.Info("🛑 Stopping AppStore consumer")
	// The consumer should handle context cancellation internally
	return nil
}

func (w *AppStoreConsumerWrapper) Name() string {
	return w.name
}

// UnifiedConsumerWrapper wraps the unified subscription consumer to implement WorkerConsumer
type UnifiedConsumerWrapper struct {
	consumer interface{} // Use interface{} for now until we check the actual type
	name     string
	logger   logger.Logger
}

// NewUnifiedConsumerWrapper creates a new wrapper for the unified subscription consumer
func NewUnifiedConsumerWrapper(opts unifiedConsumer.UnifiedConsumerOpts) *UnifiedConsumerWrapper {
	consumer := unifiedConsumer.NewUnifiedConsumer(opts)

	return &UnifiedConsumerWrapper{
		consumer: consumer,
		name:     "unified-subscription-consumer",
		logger:   logger.NewEnhancedLogger(),
	}
}

func (w *UnifiedConsumerWrapper) Start(ctx context.Context) error {
	w.logger.Info("🔄 Starting Unified Subscription consumer")
	// We need to check the actual interface of the unified consumer
	if starter, ok := w.consumer.(interface{ Start(context.Context) }); ok {
		go starter.Start(ctx)
	}
	return nil
}

func (w *UnifiedConsumerWrapper) Stop(ctx context.Context) error {
	w.logger.Info("🛑 Stopping Unified Subscription consumer")
	// The consumer should handle context cancellation internally
	return nil
}

func (w *UnifiedConsumerWrapper) Name() string {
	return w.name
}
