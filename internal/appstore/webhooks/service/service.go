package service

import (
	"context"
	"time"

	"subsnotifpro-go/internal/appstore/events"
	subscriptionSvcPkg "subsnotifpro-go/internal/appstore/subscription/service"
	"subsnotifpro-go/internal/appstore/webhooks/converter"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/appstore/webhooks/repository"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"gorm.io/gorm"
)

type AppStoreNotificationsService interface {
	ProcessNotificationForPublish(ctx context.Context, notification *dto.AppStoreNotification) error
	ProcessAppStoreEvent(ctx context.Context, id uuid.UUID) error
}

type appStoreNotificationsService struct {
	repo                repository.AppstoreNotificationsRepository
	converter           converter.NotificationConverter
	db                  *gorm.DB
	cb                  *gobreaker.CircuitBreaker
	publisher           events.EventPublisher
	subscriptionService subscriptionSvcPkg.AppStoreSubscriptionService
}

func NewAppStoreNotificationsService(
	db *gorm.DB,
	publisher events.EventPublisher,
	subscriptionService subscriptionSvcPkg.AppStoreSubscriptionService,
) AppStoreNotificationsService {
	return &appStoreNotificationsService{
		repo:      repository.NewAppstoreNotificationsRepository(db),
		converter: *converter.NewNotificationConverter(),
		db:        db,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "AppStoreNotification",
			MaxRequests: 5,
			Interval:    1 * time.Minute,
			Timeout:     15 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
		}),
		publisher:           publisher,
		subscriptionService: subscriptionService,
	}
}
