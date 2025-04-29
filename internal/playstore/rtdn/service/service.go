// internal/google_playstore/rtdn/service/service.go
package service

import (
	"context"
	"time"

	"subsnotifpro-go/config"
	apiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/events"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"
	"subsnotifpro-go/queue"

	"github.com/sony/gobreaker"
	"gorm.io/gorm"
)

// RTDNService defines an interface for RTDN service methods
type RTDNService interface {
	ProcessWebhookEventForPublish(ctx context.Context, event *dto.GooglePlayWebhookEvent) error
	ProcessWebhookEvent(ctx context.Context, payload models.GooglePublishPayload) error

	ProcessSubscriptionEvent(ctx context.Context, payload models.GooglePublishPayload) error
	ProcessOneTimeProductEvent(ctx context.Context, payload models.GooglePublishPayload) error
	ProcessVoidedPurchaseEvent(ctx context.Context, payload models.GooglePublishPayload) error

	GetDLQSize(ctx context.Context) (int, error)
	RetryMessages(ctx context.Context) error
}

// rtdnService implements the RTDNService interface

// Compile-time interface implementation check for event processor
var _ events.EventProcessor = (*rtdnService)(nil)

type rtdnService struct {
	repo                         repository.RTDNRepository
	ctx                          context.Context
	apiService                   apiService.PlaystoreApiService
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService
	db                           *gorm.DB
	cb                           *gobreaker.CircuitBreaker
	publisher                    events.EventPublisher

	rmqManager *queue.RabbitMQManager
	cfg        *config.RabbitMQConfig
}

// NewRTDNService creates a new instance of RTDNService
func NewRTDNService(ctx context.Context,
	repo repository.RTDNRepository,
	apiService apiService.PlaystoreApiService,
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService,
	db *gorm.DB,
	publisher events.EventPublisher,
	rmqManager *queue.RabbitMQManager,
	cfg *config.RabbitMQConfig,
) RTDNService {
	service := &rtdnService{repo: repo,
		ctx:                          ctx,
		apiService:                   apiService,
		playstoreSubscriptionService: playstoreSubscriptionService,
		db:                           db,

		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "PlayStoreAPI",
			MaxRequests: 5,
			Interval:    1 * time.Minute,
			Timeout:     15 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
		}),
		publisher:  publisher,
		rmqManager: rmqManager,
		cfg:        cfg,
	}

	return service
}
